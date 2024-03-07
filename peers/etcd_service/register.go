package etcdservice

import (
	"TDKCache/service/log"
	"time"

	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var logger = log.NewLogger("etcd", "Register")

// ServiceRegister 创建租约注册服务
type ServiceRegister struct {
	cli     *clientv3.Client // etcd客户端
	leaseID clientv3.LeaseID // 租约ID
	// 租约keepalive对应的通道
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	value         string
}

// NewServiceResigter 新建注册服务
func NewServiceResigter(endpoints []string, key, value string, ttl int64) (*ServiceRegister, error) {
	// 创建etcd客户端
	cli, err := clientv3.New(
		clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: 5 * time.Second,
		})
	if err != nil {
		logger.Error("get etcd client: %v", err)
		return nil, err
	}

	service := &ServiceRegister{
		cli:   cli,
		key:   key,
		value: value,
	}

	// 申请租约，设置时间keepalive
	if err := service.putKeyWithLease(ttl); err != nil {
		logger.Error("putKeyWithLease: %v", err)
		return nil, err
	}

	return service, nil
}

func (s *ServiceRegister) putKeyWithLease(ttl int64) error {
	// 创建租约
	resp, err := s.cli.Lease.Grant(context.Background(), ttl)
	if err != nil {
		logger.Error("create lease: %v", err)
		return err
	}

	_, err = s.cli.Put(context.Background(), s.key, s.value, clientv3.WithLease(clientv3.LeaseID(resp.ID)))
	if err != nil {
		logger.Error("put key: %v", err)
		return err
	}

	leaseRespChan, err := s.cli.KeepAlive(context.Background(), clientv3.LeaseID(resp.ID))

	if err != nil {
		logger.Error("get keep alive chan: %v", err)
		return err
	}

	s.leaseID = clientv3.LeaseID(resp.ID)
	s.keepAliveChan = leaseRespChan
	logger.Info("Put key with leaseID[%d]: %s -> val: %s success", resp.ID, s.key, s.value)

	return nil
}

// 监听租约的续租
func (s *ServiceRegister) ListenLeaseRespChan() {
	go func() {
		for leaseKeepResp := range s.keepAliveChan {
			logger.Debug("lease keep alive success: %v", leaseKeepResp)
		}
		logger.Info("Close lease")
	}()
}

// 注销服务
func (s *ServiceRegister) Close() error {
	// 撤销租约
	if _, err := s.cli.Revoke(context.Background(), s.leaseID); err != nil {
		logger.Error("revoke lease: %v", err)
		return err
	}
	logger.Info("Revoke lease success")
	return s.cli.Close()
}
