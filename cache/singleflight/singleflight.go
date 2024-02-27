package singleflight

import "sync"

// call是正在执行或已完成的一个请求
type call struct {
	wg  sync.WaitGroup // 使用waitgrou避免重入
	val interface{}    // 请求返回值
	err error          // 错误
	cnt int            // 计数
}

// Group包含了若干组执行中的call
type Group struct {
	mu sync.Mutex       // 加锁
	m  map[string]*call // 以key为索引的call哈希表
}

// Do方法保证针对相同的key，无论Do被调用多少次，函数fn都只会被调用一次
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	// 懒初始化
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		// 如果哈希表中已经有对key的请求进行中，则避免重入，等待已有请求的返回结果
		g.mu.Unlock()
		c.cnt++
		// 等到已有请求c完成
		c.wg.Wait()
		return c.val, c.err
	}
	// 如果没有进行中的请求，则创建
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// 运行fn函数
	c.val, c.err = fn()
	c.wg.Done()

	// 请求结束，删除哈希表中对应的call
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err

}
