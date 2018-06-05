package main

import (
	"path/filepath"

	server "github.com/kgantsov/kvgo/pkg/server"
	log "github.com/sirupsen/logrus"
)

const (
	port     = ":50051"
	raftAddr = ":12000"
)

func main() {
	log.Info("Creating storage...")
	dbPath := filepath.Join(".", "data.db")
	indexPath := filepath.Join(".", "indexes.idx")
	raftDir := filepath.Join(".", "raft")

	store := server.NewStore(dbPath, indexPath, 4, 100, raftDir, raftAddr)
	log.Info("Storage was succesfully created")

	server.ListenAndServGrpc(":50051", store)

	log.Info("Listening on port: ", port[1:])
}
