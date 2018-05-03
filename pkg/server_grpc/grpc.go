package server

import (
	"fmt"
	"net"

	kv "github.com/kgantsov/kvgo/pkg/kv"
	log "github.com/sirupsen/logrus"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

type server struct {
	store *kv.KV
}

func ListenAndServGrpc(port string, store *kv.KV) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := &server{}
	s.store = store

	grpcServer := grpc.NewServer()

	log.Info("Listening on port: ", port[1:])

	RegisterKVServer(grpcServer, s)

	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) Set(ctx context.Context, in *SetRequest) (*SetResponse, error) {
	s.store.Set(in.Key, in.Value)
	return &SetResponse{Exist: true}, nil
}

func (s *server) Get(ctx context.Context, in *GetRequest) (*GetResponse, error) {
	val, ok := s.store.Get(in.Key)
	fmt.Println(">>>>>", val, ok)
	return &GetResponse{Exist: ok, Value: val}, nil
}

func (s *server) Del(ctx context.Context, in *DelRequest) (*DelResponse, error) {
	s.store.Delete(in.Key)
	return &DelResponse{Exist: false}, nil
}