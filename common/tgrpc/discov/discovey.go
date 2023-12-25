package discov

import "context"

type Discovery interface {
	// 服务发现名字 eg etcd zk consul
	Name() string
	// 注册服务
	Register(ctx context.Context, service *Service)
	// 取消注册服务
	UnRegister(ctx context.Context, service *Service)
	// 获取服务节点信息
	GetService(ctx context.Context, name string) *Service
	// 增加监听者
	AddListener(ctx context.Context, f func())
	// 通知所有的监听者
	NotifyListeners()
}
