package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	kv "github.com/kgantsov/kvgo/pkg/kv"
	log "github.com/sirupsen/logrus"
)

type Store struct {
	RaftDir  string
	RaftBind string

	KV *kv.KV

	raft *raft.Raft // The consensus mechanism
}

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

type command struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func NewStore(dbPath, indexPath string, blockSize uint32, maxBlockNumber int16, RaftDir, RaftBind string) *Store {
	store := new(Store)
	store.KV = kv.NewKV(dbPath, indexPath, blockSize, maxBlockNumber)

	store.RaftDir = RaftDir
	store.RaftBind = RaftBind

	return store
}

// Open opens the store. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore leader, of the cluster.
// localID should be the server identifier for this node.
func (s *Store) Open(enableSingle bool, localID string) error {
	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localID)

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", s.RaftBind)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(s.RaftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(s.RaftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(s.RaftDir, "raft.db"))
	if err != nil {
		return fmt.Errorf("new bolt store: %s", err)
	}

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, (*FSM)(s), logStore, logStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.raft = ra

	if enableSingle {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	go s.Compacter()

	return nil
}

func (s *Store) Set(key, value string) {
	if s.raft.State() != raft.Leader {
		log.Error("not leader")
		return
	}

	c := &command{
		Op:    "set",
		Key:   key,
		Value: value,
	}
	b, err := json.Marshal(c)
	if err != nil {
		log.Error(err.Error())
		return
	}

	s.raft.Apply(b, raftTimeout)
}

func (s *Store) Get(key string) (string, bool) {
	val, ok := s.KV.Get(key)

	return val, ok
}

func (s *Store) Delete(key string) {
	if s.raft.State() != raft.Leader {
		log.Error("not leader")
		return
	}

	c := &command{
		Op:  "delete",
		Key: key,
	}
	b, err := json.Marshal(c)
	if err != nil {
		log.Error(err.Error())
		return
	}

	s.raft.Apply(b, raftTimeout)
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (s *Store) Join(nodeID, addr string) error {
	log.Info("Rreceived join request for remote node %s at %s", nodeID, addr)

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Info("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				log.Info("Node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("Error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	log.Info("Node %s at %s joined successfully", nodeID, addr)
	return nil
}

func (s *Store) SyncToDisk() {
	s.KV.SyncToDisk()
}

func (s *Store) CompactData() {
	s.KV.CompactData()
}

func (s *Store) Close() {
	s.KV.Close()
}

func (s *Store) Compacter() {
	tick := time.Tick(24 * time.Hour)
	for {
		select {
		case <-tick:
			s.CompactData()
			continue
		}
	}
}

type FSM Store

// Apply applies a Raft log entry to the key-value store.
func (f *FSM) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.Op {
	case "set":
		return f.applySet(c.Key, c.Value)
	case "delete":
		return f.applyDelete(c.Key)
	default:
		panic(fmt.Sprintf("unrecognized command op: %s", c.Op))
	}
}

// Snapshot returns a snapshot of the key-value store.
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.KV.Lock.Lock()
	defer f.KV.Lock.Unlock()

	// Clone the map.
	o := make(map[string]string)
	for k, v := range f.KV.MemTable {
		o[k] = v
	}
	return &fsmSnapshot{store: o}, nil
}

// Restore stores the key-value store to a previous state.
func (f *FSM) Restore(rc io.ReadCloser) error {
	o := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&o); err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	f.KV.MemTable = o
	return nil
}

func (f *FSM) applySet(key, value string) interface{} {
	f.KV.Set(key, value)
	return nil
}

func (f *FSM) applyDelete(key string) interface{} {
	f.KV.Delete(key)
	return nil
}

type fsmSnapshot struct {
	store map[string]string
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data.
		b, err := json.Marshal(f.store)
		if err != nil {
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapshot) Release() {

}
