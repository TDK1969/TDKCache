package rpc

import (
	mycache "TDKCache/cache"
	"TDKCache/peers"
	"TDKCache/service/consistenthash"
	"TDKCache/service/log"
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
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
}

var rpcLogger *log.TubeEntry

func newLogger(addr string) *log.TubeEntry {
	return log.Mylog.WithFields(logrus.Fields{
		"component": "TDKCache",
		"category":  fmt.Sprintf("RPC Server <%s>", addr),
	})
}

func NewRPCServer(addr string) *RPCServer {
	rpcLogger = newLogger(addr)
	p := &RPCServer{
		self: addr,
		addr: addr,
	}
	return p
}

func (s *RPCServer) Set(peers ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.peersMap = consistenthash.NewHashRing(nil, defaultReplicas)
	s.peersMap.Add(peers...)
	s.getters = make(map[string]*RPCGetter)

	for _, peer := range peers {
		s.getters[peer] = NewRPCGetter(peer)
	}

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

func (s *RPCServer) Start(addrs []string, g peers.GroupCache) {
	rpcLogger.Info("Start RPC server")
	s.Set(addrs...)
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
