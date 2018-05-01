package main

import (
	"path/filepath"

	kv "github.com/kgantsov/kvgo/pkg/kv"
	server_grps "github.com/kgantsov/kvgo/pkg/server_grpc"
	log "github.com/sirupsen/logrus"
)

const (
	port = ":50051"
)

func main() {
	log.Info("Creating storage...")
	dbPath := filepath.Join(".", "data.db")
	indexPath := filepath.Join(".", "indexes.idx")

	store := kv.NewKV(dbPath, indexPath, 4, 100)
	log.Info("Storage was succesfully created")

	server_grps.ListenAndServGrpc(":50051", store)

	log.Info("Listening on port: ", port[1:])
}
