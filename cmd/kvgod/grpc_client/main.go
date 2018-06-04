package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	pb "github.com/kgantsov/kvgo/pkg/server"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	rpcAddr := flag.String("rpc_addr", ":50051", "RPC bind address")
	flag.Parse()

	command := flag.Arg(0)
	key := flag.Arg(1)
	value := flag.Arg(2)

	conn, err := grpc.Dial(*rpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewKVClient(conn)

	if flag.NArg() < 2 {
		fmt.Println("Wrong number of arguments")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	switch command {
	case "get":
		r, err := c.Get(ctx, &pb.GetRequest{Key: key})
		if err != nil {
			log.Fatalf("could not get key %s: %v", key, err)
		}
		fmt.Println("Result: ", r.Value)
	case "set":
		r, err := c.Set(ctx, &pb.SetRequest{Key: key, Value: value})
		if err != nil {
			log.Fatalf("could not set key %s to a value %s: %v", key, value, err)
		}
		fmt.Println("Result: ", r.Exist)
	case "del":
		r, err := c.Del(ctx, &pb.DelRequest{Key: key})
		if err != nil {
			log.Fatalf("could not delete key %s: %v", key, err)
		}
		fmt.Println("Result: ", r.Exist)
	}
}
