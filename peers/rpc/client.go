package rpc

import (
	"TDKCache/peers/rpc/pool"
	"context"
)

// RPC通信的客户端实现
type RPCGetter struct {
	addr string
	pool pool.Pool
}

func NewRPCGetter(addr string) *RPCGetter {
	p, err := pool.NewRPCPool(addr, pool.DefaultOptions)

	if err != nil {
		panic(err)
	}

	return &RPCGetter{
		addr: addr,
		pool: p,
	}
}

func (g *RPCGetter) Get(group string, key string) ([]byte, error) {
	if g.pool == nil {
		var err error
		g.pool, err = pool.NewRPCPool(g.addr, pool.DefaultOptions)
		if err != nil {
			return nil, err
		}
	}
	// 从连接池中获取连接
	cc, err := g.pool.Get()
	if err != nil {
		return nil, err
	}
	defer cc.Close()

	c := NewPeerServiceClient(cc.Value())

	/*
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
	*/

	r, err := c.GetKey(context.Background(), &GetRequest{Group: group, Key: key})
	if err != nil {
		rpcLogger.Error("could not get key: %v", err)
		return nil, err
	}

	return r.GetValue(), nil
}
