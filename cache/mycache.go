package mycache

import (
	"TDKCache/peers"
	"TDKCache/service/log"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type Group struct {
	name      string
	getter    Getter // 缓存未命中时的回调函数
	mainCache *cache
	peers     peers.PeerPicker
}

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu          sync.RWMutex // 读写锁
	groups      = make(map[string]*Group)
	groupLogger = log.Mylog.WithFields(logrus.Fields{
		"component": "TDKCache",
		"category":  "Group",
	})
)

func NewGroup(name string, capacity int64, getter Getter) *Group {
	if getter == nil {
		groupLogger.Panic("Getter can't be nil\n")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: NewCache(capacity, nil),
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		groupLogger.Info("key [%s] hit: %v\n", key, v)
		return v, nil
	}
	groupLogger.Info("key [%s] miss\n", key)
	return g.load(key)
}

// getFromPeer get key from peer
func (g *Group) getFromPeer(peer peers.PeerGetter, key string) (ByteView, error) {
	if bytes, err := peer.Get(g.name, key); err != nil {
		groupLogger.Info("failed to get key [%s] from peer", key)
		return ByteView{}, err
	} else {
		return ByteView{data: bytes}, nil
	}
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 当key不在缓存时,从远程或本地获取需要缓存的值
	// 从远程获取
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			groupLogger.Info("failed to get from peer: %v", err)
		}
	}
	// 先从本地获取缓存
	return g.getLocally(key)

}

func (g *Group) getLocally(key string) (ByteView, error) {
	groupLogger.Info("get key [%s] locally\n", key)
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{data: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// RegisterPeers向Group注册 PeerPicker
func (g *Group) RegisterPeers(peers peers.PeerPicker) {
	if g.peers != nil {
		groupLogger.Panic("RegisterPeers called more than once")
		panic("RegisterPeers called more than once")
	}
	g.peers = peers
}
