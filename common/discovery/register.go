package discovery

import (
	"context"
	"log"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/kuan525/tiger/common/config"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// 创建租约注册服务
type ServiceRegister struct {
	cli           *clientv3.Client                        // etcd client
	leaseID       clientv3.LeaseID                        // 租约ID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse // 租约keepalieve相应chan
	key           string                                  // key
	val           string                                  // value
	ctx           *context.Context
}

// 新建注册服务
func NewServiceRegister(ctx *context.Context, key string, endPointInfo *EndpointInfo, lease int64) (*ServiceRegister, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.GetEndPointsForDiscovery(),
		DialTimeout: config.GetTimeoutForDiscovery(),
	})
	if err != nil {
		log.Fatal(err)
	}

	sevg := &ServiceRegister{
		cli: cli,
		key: key,
		val: endPointInfo.Marshal(),
		ctx: ctx,
	}

	// 申请租约设置时间keepalive
	if err := sevg.putKeyWithLease(lease); err != nil {
		return nil, err
	}
	return sevg, nil
}

// 设置租约
func (s *ServiceRegister) putKeyWithLease(lease int64) error {
	// 设置租约时间
	resp, err := s.cli.Grant(*s.ctx, lease)
	if err != nil {
		return err
	}
	// 注册服务并绑定租约
	_, err = s.cli.Put(*s.ctx, s.key, s.val, clientv3.WithLease(resp.ID))
	if err != nil {
		return nil
	}
	// 设置租约，定期发送需求请求
	leaseRespChan, err := s.cli.KeepAlive(*s.ctx, resp.ID)
	if err != nil {
		return err
	}
	s.leaseID = resp.ID
	s.keepAliveChan = leaseRespChan
	return nil
}

func (s *ServiceRegister) UpdateValue(val *EndpointInfo) error {
	value := val.Marshal()
	_, err := s.cli.Put(*s.ctx, s.key, value, clientv3.WithLease(s.leaseID))
	if err != nil {
		return err
	}
	s.val = value
	logger.CtxInfof(*s.ctx, "ServiceRegister.updateValue leaseID=%d Put key=%s val=%s, success!", s.leaseID, s.key, s.val)
	return nil
}

// 监听/续租情况
func (s *ServiceRegister) ListenLeaseRespChan() {
	for leaseKeepResp := range s.keepAliveChan {
		logger.CtxInfof(*s.ctx, "lease success leaseID:%d, Put key:%s,val:%s resp:+%v", s.leaseID, s.key, s.val, leaseKeepResp)
	}
	logger.CtxInfof(*s.ctx, "lease failed !!! leaseID:%d, Put key:%s,val:%s", s.leaseID, s.key, s.val)
}

func (s *ServiceRegister) Close() error {
	// 撤销租约
	if _, err := s.cli.Revoke(context.Background(), s.leaseID); err != nil {
		return err
	}
	logger.CtxInfof(*s.ctx, "lease close !!! leaseID:%d, Put Key:%s,val:%s success!", s.leaseID, s.key, s.val)
	return s.cli.Close()
}
