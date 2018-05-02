package main

import (
	"context"
	"fmt"
	"os"
	"time"

	pb "github.com/kgantsov/kvgo/pkg/server_grpc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewKVClient(conn)

	if len(os.Args) < 3 {
		fmt.Println("Wrong number of arguments")
	}

	command := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	switch command {
	case "get":
		fmt.Println("get")
		r, err := c.Get(ctx, &pb.GetRequest{Key: os.Args[2]})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Result: %s", r.Value)
	case "set":
		fmt.Println("set")
		r, err := c.Set(ctx, &pb.SetRequest{Key: os.Args[2], Value: os.Args[3]})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Result: %s", r.Exist)
	case "del":
		fmt.Println("del")
		r, err := c.Del(ctx, &pb.DelRequest{Key: os.Args[2]})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Result: %s", r.Exist)
	}
}
