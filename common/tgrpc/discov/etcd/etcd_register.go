package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kuan525/tiger/common/tgrpc/discov"

	"github.com/bytedance/gopkg/util/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const KeyPrefix = "tiger/tgrpc"

// Register
type Register struct {
	Options // etcd的一些信息
	cli     *clientv3.Client

	serviceRegisterCh   chan *discov.Service
	serviceUnRegisterCh chan *discov.Service

	lock             sync.Mutex
	downServices     atomic.Value
	registerServices map[string]*registerService
	listeners        []func()
}

type registerService struct {
	service      *discov.Service
	leaseID      clientv3.LeaseID
	isRegistered bool
	keepAliveCh  <-chan *clientv3.LeaseKeepAliveResponse
}

func NewETCDRegister(opts ...Option) (discov.Discovery, error) {
	opt := defaultOptions
	// 对默认options进行修改，通过若干个option函数
	for _, o := range opts {
		o(&opt)
	}

	r := &Register{
		Options:             opt,
		serviceRegisterCh:   make(chan *discov.Service),
		serviceUnRegisterCh: make(chan *discov.Service),
		lock:                sync.Mutex{},
		downServices:        atomic.Value{},
		registerServices:    make(map[string]*registerService),
	}
	if err := r.Init(context.TODO()); err != nil {
		return nil, err
	}
	return r, nil
}

// 初始化cli，同时开启监听Register中唯二的队列
func (r *Register) Init(ctx context.Context) error {
	var err error
	r.cli, err = clientv3.New(
		clientv3.Config{
			Endpoints:   r.endpoints,
			DialTimeout: r.dialTimeout,
		})
	if err != nil {
		return err
	}
	// 开启监听
	go r.run()
	return nil
}

func (r *Register) run() {
	for {
		select {
		case service := <-r.serviceRegisterCh:
			if _, ok := r.registerServices[service.Name]; ok {
				r.registerServices[service.Name].service.Endpoints = append(r.registerServices[service.Name].service.Endpoints, service.Endpoints...)
				r.registerServices[service.Name].isRegistered = false
			} else {
				r.registerServices[service.Name] = &registerService{
					service:      service,
					isRegistered: false,
				}
			}
		case service := <-r.serviceUnRegisterCh:
			if _, ok := r.registerServices[service.Name]; !ok { // 无则不处理，打个日志
				logger.CtxErrorf(context.TODO(), "UnRegisterService err, service %v was not registered", service.Name)
			} else {
				r.unRegisterService(context.TODO(), service)
			}
		default:
			// 这里轮训，全量刷新
			r.registerServiceOrKeepAlive(context.TODO())
			time.Sleep(r.registerServiceOrKeepAliveInterval)
		}
	}
}

func (r *Register) registerServiceOrKeepAlive(ctx context.Context) {
	for _, service := range r.registerServices {
		if !service.isRegistered { // 未上报
			r.registerService(ctx, service)
			r.registerServices[service.service.Name].isRegistered = true
		} else {
			// r.cli.KeepAlive返回的channel中的数据不处理，会阻塞，则不会续期
			r.KeepAlive(ctx, service)
		}
	}
}

func (r *Register) registerService(ctx context.Context, service *registerService) {
	// 获取一个租约
	leaseGrantResp, err := r.cli.Grant(ctx, r.keepAliveInterval)
	if err != nil {
		logger.CtxErrorf(ctx, "register service grant, err:%v", err)
		return
	}
	service.leaseID = leaseGrantResp.ID

	for _, endpoint := range service.service.Endpoints {
		key := r.getEtcdRegisterKey(service.service.Name, endpoint.IP, endpoint.Port)
		raw, err := json.Marshal(endpoint)
		if err != nil {
			logger.CtxErrorf(ctx, "register service err,err:%v, register data:%v", err, string(raw))
			continue
		}

		// 每一个endpoint都推上去
		_, err = r.cli.Put(ctx, key, string(raw), clientv3.WithLease(leaseGrantResp.ID))
		if err != nil {
			logger.CtxErrorf(ctx, "register service err,err:%v, register data:%v", err, string(raw))
			continue
		}
	}

	// 设置一个通道，定期发送心跳消息保持租约有效，自动续期租约
	keepAliveCh, err := r.cli.KeepAlive(ctx, leaseGrantResp.ID)
	if err != nil {
		logger.CtxErrorf(ctx, "register service keepalive,err:%v", err)
		return
	}
	service.keepAliveCh = keepAliveCh
	service.isRegistered = true
}

func (r *Register) unRegisterService(ctx context.Context, service *discov.Service) {
	endpoints := make([]*discov.Endpoint, 0)
	for _, endpoint := range r.registerServices[service.Name].service.Endpoints { // 所有endpoint
		var isRemove bool
		for _, unRegisterEndpoint := range service.Endpoints { // 要移除的endpoint
			if endpoint.IP == unRegisterEndpoint.IP && endpoint.Port == unRegisterEndpoint.Port {
				_, err := r.cli.Delete(context.TODO(), r.getEtcdRegisterKey(service.Name, endpoint.IP, endpoint.Port))
				if err != nil {
					logger.CtxErrorf(ctx, "UnRegisterService etcd del err, service %v was not registered", service.Name)
				}
				isRemove = true
				break
			}
		}

		if !isRemove {
			endpoints = append(endpoints, endpoint)
		}
	}
	if len(endpoints) == 0 {
		delete(r.registerServices, service.Name)
	} else {
		r.registerServices[service.Name].service.Endpoints = endpoints
	}
}

func (r *Register) KeepAlive(ctx context.Context, service *registerService) {
	for {
		select {
		case <-service.keepAliveCh:
		default:
			return
		}
	}
}

// downServices中没有，会进入一次，监听变化，再修改downService，通知Listeners
func (r *Register) watch(ctx context.Context, key string, revision int64) {
	rch := r.cli.Watch(ctx, key, clientv3.WithRev(revision), clientv3.WithPrefix())
	for n := range rch {
		for _, ev := range n.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				var endpoint discov.Endpoint
				if err := json.Unmarshal(ev.Kv.Value, &endpoint); err != nil {
					continue
				}
				serviceName, _, _ := r.getServiceNameByEtcdKey(string(ev.Kv.Key))
				r.updateDownService(&discov.Service{Name: serviceName, Endpoints: []*discov.Endpoint{&endpoint}})
			case clientv3.EventTypeDelete:
				var endpoint discov.Service
				if err := json.Unmarshal(ev.Kv.Value, &endpoint); err != nil {
					continue
				}
				serviceName, ip, port := r.getServiceNameByEtcdKey(string(ev.Kv.Key))
				r.delDownService(&discov.Service{
					Name: serviceName,
					Endpoints: []*discov.Endpoint{
						{
							IP:   ip,
							Port: port,
						},
					},
				})
			}
		}
	}
}

func (r *Register) updateDownService(service *discov.Service) {
	r.lock.Lock()
	defer r.lock.Unlock()

	downServices := r.downServices.Load().(map[string]*discov.Service)
	// 如果不存在，则加入，然后退出，不需要通知Listeners
	if _, ok := downServices[service.Name]; !ok {
		downServices[service.Name] = service
		r.downServices.Store(downServices)
		return
	}

	// 新的endpoint加入downservice
	for _, newAddEndpoint := range service.Endpoints {
		var isExist bool
		for idx, endpoint := range downServices[service.Name].Endpoints {
			if newAddEndpoint.IP == endpoint.IP && newAddEndpoint.Port == endpoint.Port {
				downServices[service.Name].Endpoints[idx] = newAddEndpoint // 更新
				isExist = true                                             // 标记存在
				break
			}
		}
		if !isExist { // 如果不存在，则加入
			downServices[service.Name].Endpoints = append(downServices[service.Name].Endpoints, newAddEndpoint)
		}
	}
	r.downServices.Store(downServices)

	// 通知Listeners，更新一下
	r.NotifyListeners()
}

func (r *Register) delDownService(service *discov.Service) {
	r.lock.Lock()
	defer r.lock.Unlock()

	downServices := r.downServices.Load().(map[string]*discov.Service)
	if _, ok := downServices[service.Name]; !ok {
		return
	}
	endpoints := make([]*discov.Endpoint, 0)
	for _, endpoint := range downServices[service.Name].Endpoints {
		var isRemove bool
		for _, delEndpoint := range service.Endpoints {
			if delEndpoint.IP == endpoint.IP && delEndpoint.Port == endpoint.Port {
				isRemove = true
				break
			}
		}
		if !isRemove {
			endpoints = append(endpoints, endpoint)
		}
	}
	downServices[service.Name].Endpoints = endpoints
	r.downServices.Store(downServices)

	// 修改之后，通知Listeners
	r.NotifyListeners()
}

func (r *Register) getDownServices() map[string]*discov.Service {
	allServices := r.downServices.Load()
	if allServices == nil {
		return make(map[string]*discov.Service, 0)
	}
	return allServices.(map[string]*discov.Service)
}

func (r *Register) getEtcdRegisterKey(name, ip string, port int) string {
	return fmt.Sprintf(KeyPrefix+"%s/%s/%d", name, ip, port)
}

func (r *Register) getEtcdRegisterPrefixKey(name string) string {
	return fmt.Sprintf(KeyPrefix+"%s", name)
}

// name ip port
func (r *Register) getServiceNameByEtcdKey(key string) (string, string, int) {
	trimStr := strings.TrimPrefix(key, KeyPrefix)
	strs := strings.Split(trimStr, "/")

	port, _ := strconv.Atoi(strs[2])
	return strs[0], strs[1], port
}

// ----------------------------------------------
// 以下Register实现Discovery
// ----------------------------------------------
func (r *Register) Name() string {
	return "etcd"
}

func (r *Register) Register(ctx context.Context, service *discov.Service) {
	r.serviceRegisterCh <- service
}

func (r *Register) UnRegister(ctx context.Context, service *discov.Service) {
	r.serviceUnRegisterCh <- service
}

// 这里会更新一下 downServices
func (r *Register) GetService(ctx context.Context, name string) *discov.Service {
	allServices := r.getDownServices()
	// 先从本地拉，再从etcd拉
	if val, ok := allServices[name]; ok {
		return val
	}

	// 防止并发获取service导致cache中的数据混乱
	r.lock.Lock()
	defer r.lock.Unlock()

	key := r.getEtcdRegisterPrefixKey(name)
	getResp, _ := r.cli.Get(ctx, key, clientv3.WithPrefix())
	service := &discov.Service{
		Name:      name,
		Endpoints: make([]*discov.Endpoint, 0),
	}

	for _, item := range getResp.Kvs {
		var endpoint discov.Endpoint
		if err := json.Unmarshal(item.Value, &endpoint); err != nil {
			continue
		}
		service.Endpoints = append(service.Endpoints, &endpoint)
	}

	allServices[name] = service
	r.downServices.Store(allServices)

	go r.watch(ctx, key, getResp.Header.Revision+1)

	return service
}

func (r *Register) AddListener(ctx context.Context, f func()) {
	r.listeners = append(r.listeners, f)
}

func (r *Register) NotifyListeners() {
	for _, listener := range r.listeners {
		listener()
	}
}
