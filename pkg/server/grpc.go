package server

import (
	"net"

	log "github.com/sirupsen/logrus"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

type server struct {
	store *Store
}

func ListenAndServGrpc(port string, store *Store) {
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
	val, err := s.store.Get(in.Key)
	if err == nil {
		return &GetResponse{Exist: true, Value: val}, nil
	} else {
		return &GetResponse{Exist: false, Value: ""}, nil
	}
}

func (s *server) Del(ctx context.Context, in *DelRequest) (*DelResponse, error) {
	err := s.store.Delete(in.Key)
	if err == nil {
		return &DelResponse{Exist: false}, nil
	} else {
		return &DelResponse{Exist: true}, nil
	}
}

func (s *server) Join(ctx context.Context, in *JoinRequest) (*JoinResponse, error) {
	s.store.Join(in.NodeID, in.Addr)
	return &JoinResponse{Joined: true}, nil
}
