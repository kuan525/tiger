package tgrpc

import (
	"context"
	"fmt"
	"time"

	"github.com/kuan525/tiger/common/tgrpc/discov/plugin"

	"google.golang.org/grpc/resolver"

	"github.com/kuan525/tiger/common/tgrpc/discov"
	clientinterceptor "github.com/kuan525/tiger/common/tgrpc/interceptor/client"
	presolver "github.com/kuan525/tiger/common/tgrpc/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

const (
	dialTimeout = 5 * time.Second
)

type PClient struct {
	serviceName  string
	d            discov.Discovery
	interceptors []grpc.UnaryClientInterceptor
	conn         *grpc.ClientConn
}

func NewClient(serviceName string, interceptors ...grpc.UnaryClientInterceptor) (*PClient, error) {
	p := &PClient{
		serviceName:  serviceName,
		interceptors: interceptors,
	}

	if p.d == nil {
		dis, err := plugin.GetDiscovInstance()
		if err != nil {
			panic(err)
		}

		p.d = dis
	}

	// 注册名称解析器
	resolver.Register(presolver.NewDiscovBuilder(p.d))

	conn, err := p.dial()
	p.conn = conn

	return p, err
}

func (p *PClient) Conn() *grpc.ClientConn {
	return p.conn
}

func (p *PClient) dial() (*grpc.ClientConn, error) {
	// Round Robin负载均衡策略
	svcCfg := fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, roundrobin.Name)
	balancerOpt := grpc.WithDefaultServiceConfig(svcCfg)

	interceptors := []grpc.UnaryClientInterceptor{
		clientinterceptor.TraceUnaryClientInterceptor(),
		clientinterceptor.MetricUnaryClientInterceptor(),
	}
	interceptors = append(interceptors, p.interceptors...)

	options := []grpc.DialOption{
		balancerOpt,
		grpc.WithChainUnaryInterceptor(interceptors...),
		grpc.WithInsecure(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	return grpc.DialContext(ctx, fmt.Sprintf("discov:///%v", p.serviceName), options...)
}

func (p *PClient) DialByEndPoint(adrss string) (*grpc.ClientConn, error) {
	interceptors := []grpc.UnaryClientInterceptor{
		clientinterceptor.TraceUnaryClientInterceptor(),
		clientinterceptor.MetricUnaryClientInterceptor(),
	}
	interceptors = append(interceptors, p.interceptors...)

	options := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(interceptors...),
		grpc.WithInsecure(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	return grpc.DialContext(ctx, adrss, options...)
}
