package discovery

import (
	"context"
	"sync"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/kuan525/tiger/common/config"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// 服务发现
type ServiceDiscovery struct {
	cli  *clientv3.Client // etcd client
	lock sync.Mutex
	ctx  *context.Context
}

// 新建服务发现
func NewServiceDiscovery(ctx *context.Context) *ServiceDiscovery {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.GetEndPointsForDiscovery(),
		DialTimeout: config.GetTimeoutForDiscovery(),
	})
	if err != nil {
		logger.Fatal(err)
	}
	return &ServiceDiscovery{
		cli: cli,
		ctx: ctx,
	}
}

// 初始化服务列表和监视
func (s *ServiceDiscovery) WatchService(prefix string, set, del func(key, value string)) error {
	// 根据前缀获取现有的key
	resp, err := s.cli.Get(*s.ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, ev := range resp.Kvs {
		set(string(ev.Key), string(ev.Value))
	}
	// 监视前缀，修改变更的server
	s.watcher(prefix, resp.Header.Revision+1, set, del)
	return nil
}

// 监听前缀
func (s *ServiceDiscovery) watcher(prefix string, rev int64, set, del func(key, value string)) {
	rch := s.cli.Watch(*s.ctx, prefix, clientv3.WithPrefix(), clientv3.WithRev(rev))
	logger.CtxInfof(*s.ctx, "watching prefix:%s now...", prefix)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT: // 修改或者新增
				set(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE: // 删除
				del(string(ev.Kv.Key), string(ev.Kv.Value))
			}
		}
	}
}

func (s *ServiceDiscovery) Close() error {
	return s.cli.Close()
}
