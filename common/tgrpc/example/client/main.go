package main

import (
	"context"
	"log"
	"time"

	pb "github.com/kuan525/tiger/common/tgrpc/example/helloservice"
	tgresolver "github.com/kuan525/tiger/common/tgrpc/example/resolver"

	"google.golang.org/grpc"
)

const (
	address = "localhost:8080"
)

func main() {
	tgresolver.Init()
	var conn *grpc.ClientConn
	// conn, err := grpc.Dial(address, grpc.WithInsecure())
	conn, err := grpc.Dial(
		"discov:///kuan525",
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)
	ctx, cannel := context.WithTimeout(context.Background(), time.Second*2)
	defer cannel()

	response, err := c.SayHello(ctx, &pb.HelloRequest{Name: "kuan"})
	if err != nil {
		log.Fatalf("Error when calling SayHello: %s", err)
	}
	log.Printf("Response from server: %v", response.Message)
}
