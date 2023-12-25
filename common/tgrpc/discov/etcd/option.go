package etcd

import "time"

var defaultOptions = Options{
	endpoints:              []string{"127.0.0.1:2379"},
	dialTimeout:            10 * time.Second,
	syncFlushCacheInterval: 10 * time.Second,
	keepAliveInterval:      10,
}

type Options struct {
	syncFlushCacheInterval             time.Duration
	endpoints                          []string
	dialTimeout                        time.Duration
	keepAliveInterval                  int64
	registerServiceOrKeepAliveInterval time.Duration
}

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
