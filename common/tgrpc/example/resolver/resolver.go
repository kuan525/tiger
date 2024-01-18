package resolver

import (
	"google.golang.org/grpc/resolver"
)

// 自定义name resolver

const (
	myScheme   = "discov"
	myEndpoint = "kuan525"
)

// 多个，选择一个
var addrs = []string{"localhost:8080", "localhost:8081"}

// kuanResolver 自定义name resolver，实现Resolver接口
type kuanResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (r *kuanResolver) ResolveNow(o resolver.ResolveNowOptions) {
	addrStrs := r.addrsStore[r.target.Endpoint()]
	addrList := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrList[i] = resolver.Address{Addr: s}
	}
	r.cc.UpdateState(resolver.State{Addresses: addrList})
}

func (*kuanResolver) Close() {}

// kuanResolverBuilder 需实现 Builder 接口
type kuanResolverBuilder struct{}

func (*kuanResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &kuanResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			myEndpoint: addrs,
		},
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}
func (*kuanResolverBuilder) Scheme() string { return myScheme }

func Init() {
	// 注册 kuanResolverBuilder
	resolver.Register(&kuanResolverBuilder{})
}
