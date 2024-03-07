package etcdservice

import (
	"TDKCache/service/log"
	"context"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var dLogger = log.NewLogger("etcd", "Discovery")

// 定义服务发现时的回调函数
type ServiceSetCallBackFunc func(service string)

// 定义服务删除时的回调函数
type ServiceDelCallBackFunc func(service string)

// ServiceDiscovery 服务发现
type ServiceDiscovery struct {
	// etcd客户端
	cli *clientv3.Client
	// 服务表
	serviceMap map[string]string
	// 锁
	lck sync.Mutex
}

func NewServiceDiscovery(endpoints []string) (*ServiceDiscovery, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		dLogger.Error("get etcd client: %v", err)
		return nil, err
	}

	return &ServiceDiscovery{
		cli:        cli,
		serviceMap: make(map[string]string),
	}, nil
}

func (s *ServiceDiscovery) WatchService(prefix string, setFn ServiceSetCallBackFunc, delFn ServiceDelCallBackFunc) error {
	// 获取现有key
	resp, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		dLogger.Error("get service: %v", err)
		return err
	}
	for _, ev := range resp.Kvs {
		s.SetService(string(ev.Key), string(ev.Value), setFn)
	}

	// 监听前缀的改变
	go s.watcher(prefix, setFn, delFn)
	return nil
}

// watcher 监听前缀
func (s *ServiceDiscovery) watcher(prefix string, setFn ServiceSetCallBackFunc, delFn ServiceDelCallBackFunc) {
	rch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	dLogger.Info("Watching prefix: %s", prefix)
	// 监听通道
	for resp := range rch {
		for _, ev := range resp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				s.SetService(string(ev.Kv.Key), string(ev.Kv.Value), setFn)
			case mvccpb.DELETE:
				s.DeleteServie(string(ev.Kv.Key), delFn)
			}
		}
	}
}

// SetService 新增/修改服务
func (s *ServiceDiscovery) SetService(key, value string, fn ServiceSetCallBackFunc) {
	s.lck.Lock()
	defer s.lck.Unlock()
	s.serviceMap[key] = value
	if fn != nil {
		fn(value)
	}
	dLogger.Info("set service key: %s -> value: %s", key, value)
}

// DeleteServie 删除服务
func (s *ServiceDiscovery) DeleteServie(key string, fn ServiceDelCallBackFunc) {
	s.lck.Lock()
	defer s.lck.Unlock()
	if fn != nil {
		fn(s.serviceMap[key])
	}
	delete(s.serviceMap, key)
	dLogger.Info("delete service key: %s", key)
}

// GetServices 获取服务地址
func (s *ServiceDiscovery) GetServices() []string {
	s.lck.Lock()
	defer s.lck.Unlock()
	addrs := make([]string, 0)

	for _, v := range s.serviceMap {
		addrs = append(addrs, v)
	}
	return addrs
}

// Close 关闭服务
func (s *ServiceDiscovery) Close() error {
	return s.cli.Close()
}
