package mycache

import (
	"TDKCache/cache/lru"
	"TDKCache/service/log"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const deleteChanCap = 100

const expireTime = time.Minute * 10 / 1e9

//const expireTime = (time.Second * 2) / 1e9

var cacheLogger = log.Mylog.WithFields(logrus.Fields{
	"component": "TDKCache",
	"category":  "Cache",
})

// 进行并发读写的封装
type cache struct {
	lck      sync.Mutex   // 并发锁
	lru      *lru.HCCache // lru缓存
	cacheCap int64        // 缓存容量
	exMap    *exprireMap  // 记录过期键的哈希表
}

type exprireMap struct {
	timeMap      map[int64]map[string]struct{}
	keyExpireMap map[string]int64
	lck          sync.Mutex
	stopChan     chan struct{}
}

type deleteMsg struct {
	keys []string
}

func NewExprireMap() *exprireMap {
	return &exprireMap{
		timeMap:      make(map[int64]map[string]struct{}),
		keyExpireMap: make(map[string]int64),
		lck:          sync.Mutex{},
		stopChan:     make(chan struct{}),
	}
}

func NewCache(capacity int64, onEvicted func(key string, value lru.Value)) *cache {
	c := &cache{
		lck:      sync.Mutex{},
		lru:      lru.NewHCCache(capacity, onEvicted),
		cacheCap: capacity,
		exMap:    NewExprireMap(),
	}
	go c.run(time.Now().Unix())
	return c

}

func (c *cache) run(start int64) {
	t := time.NewTicker(time.Second * 1)
	defer t.Stop()

	deleteChan := make(chan *deleteMsg, deleteChanCap)

	go func() {
		for v := range deleteChan {
			c.multiDelete(v.keys)
		}
	}()

	for {
		select {
		case <-t.C:
			start++
			if len(c.exMap.timeMap[start]) > 0 {
				keys := make([]string, 0, len(c.exMap.timeMap[start]))
				for k := range c.exMap.timeMap[start] {
					keys = append(keys, k)
				}
				cacheLogger.Debug("keys [%v] expire at %d", keys, start)

				deleteChan <- &deleteMsg{keys: keys}
			}
		case <-c.exMap.stopChan:
			close(deleteChan)
			return
		}
	}

}

func (c *cache) add(key string, value ByteView) bool {
	c.lck.Lock()
	defer c.lck.Unlock()
	t := time.Now().Unix()
	if c.lru == nil {
		c.lru = lru.NewHCCache(c.cacheCap, nil)
	}
	c.exMap.lck.Lock()
	defer c.exMap.lck.Unlock()
	if exTime, ok := c.exMap.keyExpireMap[key]; ok {
		delete(c.exMap.timeMap[exTime], key)
	}

	c.exMap.keyExpireMap[key] = t + int64(expireTime)
	keyMap, ok := c.exMap.timeMap[t+int64(expireTime)]
	if !ok {
		// 如果 map 不存在，进行初始化
		keyMap = make(map[string]struct{})
		c.exMap.timeMap[t+int64(expireTime)] = keyMap
	}

	// 添加元素到内层 map
	keyMap[key] = struct{}{}
	return c.lru.Add(key, value, t)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.lck.Lock()
	defer c.lck.Unlock()
	if c.lru == nil {
		return
	}
	t := time.Now().Unix()
	cacheLogger.Debug("get key [%s] at %d\n", key, t)
	c.exMap.lck.Lock()
	defer c.exMap.lck.Unlock()
	if exTime, ok := c.exMap.keyExpireMap[key]; ok {
		delete(c.exMap.timeMap[exTime], key)
	}
	c.exMap.keyExpireMap[key] = t + int64(expireTime)
	cacheLogger.Debug("key [%s] will expire at %d\n", key, t+int64(expireTime))
	keyMap, ok := c.exMap.timeMap[t+int64(expireTime)]
	if !ok {
		// 如果 map 不存在，进行初始化
		keyMap = make(map[string]struct{})
		c.exMap.timeMap[t+int64(expireTime)] = keyMap
	}

	// 添加元素到内层 map
	keyMap[key] = struct{}{}
	cacheLogger.Debug("tring get key [%s] from lru\n", key)
	if v, ok := c.lru.Get(key, t); ok {
		return v.(ByteView), ok
	}
	cacheLogger.Debug("key [%s] miss\n", key)
	return ByteView{}, false
}

func (c *cache) delete(key string) {
	c.lck.Lock()
	defer c.lck.Unlock()
	if c.lru == nil {
		return
	}
	c.exMap.lck.Lock()
	defer c.exMap.lck.Unlock()
	if exTime, ok := c.exMap.keyExpireMap[key]; ok {
		delete(c.exMap.timeMap[exTime], key)
		delete(c.exMap.keyExpireMap, key)
	}
	c.lru.Delete(key)
}

func (c *cache) multiDelete(keys []string) {
	cacheLogger.Debug("start to delete keys [%v]\n", keys)
	c.lck.Lock()
	defer c.lck.Unlock()
	if c.lru == nil {
		return
	}
	c.exMap.lck.Lock()
	defer c.exMap.lck.Unlock()
	for _, key := range keys {
		if exTime, ok := c.exMap.keyExpireMap[key]; ok {
			delete(c.exMap.timeMap[exTime], key)
			delete(c.exMap.keyExpireMap, key)
		}
		c.lru.Delete(key)
	}
	cacheLogger.Debug("keys [%v] deleted\n", keys)

}
