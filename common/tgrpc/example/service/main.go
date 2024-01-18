package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/kuan525/tiger/common/tgrpc/example/helloservice"

	"google.golang.org/grpc"
)

const (
	address = "localhost:8080"
)

type HelloServer struct {
	pb.UnimplementedGreeterServer
}

func (s *HelloServer) SayHello(ctx context.Context, re *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: "hello " + re.Name,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", address) // 监听端口号，可以根据实际情况修改
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &HelloServer{})

	fmt.Printf("service启动成功\n")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
