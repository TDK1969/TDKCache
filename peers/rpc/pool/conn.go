package pool

import "google.golang.org/grpc"

// Conn 单个gRPC连接的接口
type Conn interface {
	// Value 返回真实的gRPC连接 type *grpc.ClientConn.
	Value() *grpc.ClientConn

	// Close 当pool未满时，减少reference而不是关闭连接
	// Close decrease the reference of grpc connection, instead of close it.
	// if the pool is full, just close it.
	Close() error
}

// conn 对grpc.ClientConn进行包装以实现Conn接口
type conn struct {
	cc   *grpc.ClientConn
	pool *pool
	once bool
}

func (c *conn) Value() *grpc.ClientConn {
	return c.cc
}

func (c *conn) Close() error {
	c.pool.decrRef()
	if c.once {
		return c.reset()
	}
	return nil
}

// 如果是一次性的连接，需要关闭
func (c *conn) reset() error {
	cc := c.cc
	c.cc = nil
	c.once = false
	if cc != nil {
		return cc.Close()
	}
	return nil
}

// 将grpc.ClientConn包装成conn
func (p *pool) wrapConn(cc *grpc.ClientConn, once bool) *conn {
	return &conn{
		cc:   cc,
		pool: p,
		once: once,
	}
}
