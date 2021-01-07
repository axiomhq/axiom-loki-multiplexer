package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/grafana/loki/pkg/logproto"
	"google.golang.org/grpc"
)

const (
	port = "0.0.0.0:3101"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

func (s *server) Push(ctx context.Context, in *pb.PushRequest) (*pb.PushResponse, error) {
	fmt.Println(">>>")
	for i, stream := range in.Streams {
		fmt.Println(i, stream)
	}
	log.Printf("Received: %v", in.Streams)
	return &pb.PushResponse{}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPusherServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
