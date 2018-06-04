package main

import (
	"context"
	"flag"
	"path/filepath"
	"time"

	pb "github.com/kgantsov/kvgo/pkg/server"
	server "github.com/kgantsov/kvgo/pkg/server"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const dbPath = "./data.db"
const indexPath = "./indexes.idx"

func main() {
	addr := flag.String("addr", ":56379", "Redis bind address")
	rpcAddr := flag.String("rpc_addr", ":50051", "RPC bind address")
	raftDir := flag.String("raft_dir", "", "RPC DB port")
	raftAddr := flag.String("raft_addr", ":12000", "Raft bind address")
	joinAddr := flag.String("join_addr", "", "Join address")
	nodeID := flag.String("node_id", "", "Node ID")
	logLevel := flag.String("log_level", "info", "Log level")
	flag.Parse()

	level, err := log.ParseLevel(*logLevel)

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	if err != nil {
		log.Fatal("Fatal error: ", err.Error())
	}
	log.SetLevel(level)

	if *raftDir == "" {
		log.Fatal("No Raft storage directory specified\n")
	}
	if *nodeID == "" {
		log.Fatal("No nodeID storage directory specified\n")
	}

	log.Info("Creating storage...")
	store := server.NewStore(
		filepath.Join(*raftDir, dbPath), filepath.Join(*raftDir, indexPath), 1000, 10000,
	)
	store.RaftDir = *raftDir
	store.RaftBind = *raftAddr

	if err := store.Open(*joinAddr == "", *nodeID); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	if *joinAddr != "" {
		conn, err := grpc.Dial(*joinAddr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewKVClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err = c.Join(ctx, &pb.JoinRequest{Addr: *raftAddr, NodeID: *nodeID})
		if err != nil {
			log.Fatalf("could not join server: %v", err)
		}
	}

	go server.ListenAndServ(*addr, store)
	server.ListenAndServGrpc(*rpcAddr, store)
}
