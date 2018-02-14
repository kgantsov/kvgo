package main

import (
	"flag"

	server "github.com/kgantsov/kvgo/pkg/server"
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
	server.ListenAndServ(port, dbPath, indexPath)
}
