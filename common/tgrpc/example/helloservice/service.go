package helloservice

import "context"

type Service struct {
}

type HelloServer struct {
}

func (s HelloServer) SayHello(ctx context.Context, re *HelloRequest) (*HelloReply, error) {
	return &HelloReply{
		Message: "hello" + re.Name,
	}, nil
}
