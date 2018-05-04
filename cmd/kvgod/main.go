package main

import (
	"flag"

	"github.com/kgantsov/kvgo/pkg/kv"
	server "github.com/kgantsov/kvgo/pkg/server"
	server_grpc "github.com/kgantsov/kvgo/pkg/server_grpc"
	log "github.com/sirupsen/logrus"
)

const dbPath = "./data.db"
const indexPath = "./indexes.idx"

func main() {
	logLevel := flag.String("log_level", "info", "Log level")
	flag.Parse()

	level, err := log.ParseLevel(*logLevel)

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	if err != nil {
		log.Fatal("Fatal error: ", err.Error())
	}
	log.SetLevel(level)

	port := ":56379"

	log.Info("Creating storage...")
	store := kv.NewKV(dbPath, indexPath, 1000, 10000)

	server.ListenAndServ(port, store)
	server_grpc.ListenAndServGrpc(":50051", store)
}
