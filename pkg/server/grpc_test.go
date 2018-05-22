package server

import (
	"context"
	fmt "fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func TestGRPCServerBasic(t *testing.T) {
	port := ":50051"
	address := "localhost" + port

	tmpDir, _ := ioutil.TempDir("", "kvgo_grpc_tests")
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(".", "data.db")
	indexPath := filepath.Join(".", "indexes.idx")
	raftDir := filepath.Join(".", "raft")

	log.Info("Creating storage...")
	store := NewStore(dbPath, indexPath, 1000, 10000)
	store.RaftDir = raftDir
	store.RaftBind = ":12000"

	if err := store.Open(true, "node1"); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	go func() {
		ListenAndServGrpc(port, store)
	}()

	time.Sleep(3 * time.Second)

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NewKVClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		_, err := c.Get(ctx, &GetRequest{Key: key})

		if err != nil {
			t.Errorf("Expected `nil`. Got `%v`\n", err)
		}

		setResp, err := c.Set(ctx, &SetRequest{Key: key, Value: value})

		if setResp.Exist != true {
			t.Errorf("Expected `true`. Got `%v`\n", setResp.Exist)
		}

		time.Sleep(500 * time.Millisecond)

		getResp, err := c.Get(ctx, &GetRequest{Key: key})

		if err != nil {
			t.Errorf("Expected `nil`. Got `%v`\n", err)
		}
		if getResp.Value != value {
			t.Errorf("Expected `%s`. Got `%v`\n", value, getResp.Value)
		}
	}
}
