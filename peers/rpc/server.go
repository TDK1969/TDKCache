package rpc

import (
	mycache "TDKCache/cache"
	"TDKCache/peers"
	etcdservice "TDKCache/peers/etcd_service"
	"TDKCache/service/conf"
	"TDKCache/service/consistenthash"
	"TDKCache/service/log"
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultReplicas = 50
)

// RPC通信的服务端实现
type RPCServer struct {
	self     string
	addr     string
	mu       sync.Mutex
	peersMap *consistenthash.HashRing
	getters  map[string]*RPCGetter
	// 通过匿名字段内嵌结构体实现继承
	UnimplementedPeerServiceServer
	register  *etcdservice.ServiceRegister
	discovery *etcdservice.ServiceDiscovery
}

var rpcLogger *log.LogEntry

func NewRPCServer(addr string) *RPCServer {
	rpcLogger = log.NewLogger("RPC Server", fmt.Sprintf("Server <%s>", addr))
	s := &RPCServer{
		self:     addr,
		addr:     addr,
		peersMap: consistenthash.NewHashRing(nil, defaultReplicas),
		getters:  make(map[string]*RPCGetter),
	}
	return s
}

func (s *RPCServer) Set(peer string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.peersMap.Add(peer)
	s.getters[peer] = NewRPCGetter(peer)

}

func (s *RPCServer) Del(peer string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.peersMap.Del(peer)
	delete(s.getters, peer)
}

func (s *RPCServer) PickPeer(key string) (peers.PeerGetter, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if peer := s.peersMap.Get(key); peer != "" && peer != s.self {
		rpcLogger.Info("pick peer %s", peer)
		return s.getters[peer], true
	}
	return nil, false
}

func (s *RPCServer) listenAndServe() {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		rpcLogger.Error("failed to listen: %v", err)
		return
	}
	server := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(
			keepalive.EnforcementPolicy{
				MinTime:             5 * time.Second,
				PermitWithoutStream: true,
			}),
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				Time:    10 * time.Second,
				Timeout: 3 * time.Second,
			},
		),
	)
	RegisterPeerServiceServer(server, s)

	err = server.Serve(lis)
	if err != nil {
		rpcLogger.Error("failed to start RPC server: %v", err)
		return
	}
}

func (s *RPCServer) Start(g peers.GroupCache) {
	rpcLogger.Info("Start RPC server")

	// 进行服务发现
	if s.discovery == nil {
		discovery, err := etcdservice.NewServiceDiscovery(
			[]string{conf.Conf.GetString("etcd.endpoints")},
		)
		if err != nil {
			rpcLogger.Panic("new service discovery: %v", err)
		}
		s.discovery = discovery

		//s.discovery.GetServices()
	}
	s.discovery.WatchService(
		conf.Conf.GetString("etcd.servicePrefix"),
		s.Set,
		s.Del,
	)

	if s.register == nil {
		register, err := etcdservice.NewServiceResigter(
			[]string{conf.Conf.GetString("etcd.endpoints")},
			conf.Conf.GetString("etcd.servicePrefix")+s.addr,
			s.addr,
			conf.Conf.GetInt64("etcd.ttl"),
		)
		if err != nil {
			rpcLogger.Panic("new service register: %v", err)
		}
		s.register = register
	}
	s.register.ListenLeaseRespChan()

	g.RegisterPeers(s)
	s.listenAndServe()
}

func (s *RPCServer) GetKey(ctx context.Context, in *GetRequest) (*GetResponse, error) {
	groupName := in.GetGroup()
	if groupName == "" {
		rpcLogger.Error("lack of necessary param [group]")
		return nil, fmt.Errorf("lack of necessary param [group]")
	}

	key := in.GetKey()
	if key == "" {
		rpcLogger.Error("lack of necessary param [key]")
		return nil, fmt.Errorf("lack of necessary param [key]")
	}

	group := mycache.GetGroup(groupName)
	if group == nil {
		rpcLogger.Error("no such group: %s", groupName)
		return nil, fmt.Errorf("no such group: %s", groupName)
	}

	view, err := group.Get(key)
	if err != nil {
		rpcLogger.Error("Internal error: %v", err)
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return &GetResponse{Value: view.ByteSlice()}, nil
}
