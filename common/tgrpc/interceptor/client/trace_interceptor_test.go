package client

import (
	"context"
	"testing"

	ttrace "github.com/kuan525/tiger/common/tgrpc/trace"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

func TestUnaryTraceInterceptor(t *testing.T) {
	ttrace.StartAgent()
	cc := new(grpc.ClientConn)
	TraceUnaryClientInterceptor()(context.TODO(), "/create", nil, nil, cc,
		func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
			opts ...grpc.CallOption) error {
			return nil
		})

	TraceUnaryClientInterceptor()(context.TODO(), "/update", nil, nil, cc,
		func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
			opts ...grpc.CallOption) error {
			return status.Error(codes.DataLoss, "dummy")
		})

	defer ttrace.StopAgent()
}
