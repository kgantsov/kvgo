package main

import (
	"flag"
	"fmt"

	server "github.com/kgantsov/kvgo/pkg/server"
	log "github.com/sirupsen/logrus"
)

const dbPath = "./data.db"
const indexPath = "./indexes.idx"

func main() {
	port := flag.String("port", "56379", "DB port")
	rpcPort := flag.String("rpc_port", "50051", "RPC DB port")
	logLevel := flag.String("log_level", "info", "Log level")
	flag.Parse()

	level, err := log.ParseLevel(*logLevel)

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	if err != nil {
		log.Fatal("Fatal error: ", err.Error())
	}
	log.SetLevel(level)

	log.Info("Creating storage...")
	store := server.NewStore(dbPath, indexPath, 1000, 10000)

	server.ListenAndServ(fmt.Sprintf(":%s", *port), store)
	server.ListenAndServGrpc(fmt.Sprintf(":%s", *rpcPort), store)
}
