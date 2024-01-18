package etcd

import "time"

type Options struct {
	syncFlushCacheInterval             time.Duration
	dialTimeout                        time.Duration
	registerServiceOrKeepAliveInterval time.Duration
	endpoints                          []string
	keepAliveInterval                  int64
}

var defaultOptions = Options{
	endpoints:              []string{"127.0.0.1:2379"},
	dialTimeout:            10 * time.Second,
	syncFlushCacheInterval: 10 * time.Second,
	keepAliveInterval:      10,
}

// 定义一系列修改options的方法，作为插销，后续一个个运行
type Option func(o *Options)

func WithEndpoints(endpoints []string) Option {
	return func(o *Options) {
		o.endpoints = endpoints
	}
}

func WithDialTimeout(dialTimeout time.Duration) Option {
	return func(o *Options) {
		o.dialTimeout = dialTimeout
	}
}

func WithSyncFlushCacheInterval(t time.Duration) Option {
	return func(o *Options) {
		o.syncFlushCacheInterval = t
	}
}

func WithKeepAliveInterval(ttl int64) Option {
	return func(o *Options) {
		o.keepAliveInterval = ttl
	}
}

func WithRegisterServiceOrKeepAliveInterval(t time.Duration) Option {
	return func(o *Options) {
		o.registerServiceOrKeepAliveInterval = t
	}
}
