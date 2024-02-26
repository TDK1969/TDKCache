package consistenthash

import (
	"TDKCache/service/log"
	"fmt"
	"hash/crc32"
	"sort"
	"sync"

	"github.com/sirupsen/logrus"
)

// 定义哈希函数,将[]byte映射到int32
type Hash func(data []byte) uint32

// 定义哈希环
type HashRing struct {
	// 哈希函数
	hash Hash
	// 每个真实节点对应的虚拟节点
	replicas int
	// 哈希环,需要排序
	ring []uint32
	// 节点哈希值到真实节点的映射
	hash2node map[uint32]string
	// 记录所有真实节点
	nodes map[string]struct{}
	// 日志
	logger *log.TubeEntry
	// 读写锁
	lck sync.RWMutex
}

// NewHashRing 返回HashRing的指针
func NewHashRing(hash Hash, replicas int) *HashRing {
	ring := &HashRing{
		hash:      hash,
		replicas:  replicas,
		ring:      make([]uint32, 0),
		hash2node: make(map[uint32]string),
		nodes:     make(map[string]struct{}),
		logger: log.Mylog.WithFields(logrus.Fields{
			"component": "TDKCache",
			"category":  "Consistent Hash",
		}),
		lck: sync.RWMutex{},
	}
	if ring.hash == nil {
		ring.hash = crc32.ChecksumIEEE
	}
	return ring
}

// IsEmpty 返回哈希环是否为空
func (r *HashRing) IsEmpty() bool {
	return len(r.ring) == 0
}

// Add 向哈希环中加入若干节点
func (r *HashRing) Add(nodes ...string) {
	// 加写锁
	r.lck.Lock()
	defer r.lck.Unlock()
	for _, node := range nodes {
		// 如果节点已经存在,则不进行操作
		if _, ok := r.nodes[node]; ok {
			r.logger.Info("node [%s] already exists", node)
			continue
		}
		// 为每个真实节点创建虚拟节点
		for i := 0; i < r.replicas; i++ {
			hash := r.hash([]byte(fmt.Sprintf("%d%s", i, node)))
			r.ring = append(r.ring, hash)
			r.hash2node[hash] = node
			r.logger.Debug("add vitural node [%d%s] - hash [%d]", i, node, hash)
		}
		r.nodes[node] = struct{}{}
		r.logger.Info("add node [%s] successfully", node)
	}

	// 对哈希环排序
	sort.Slice(r.ring, func(i, j int) bool { return r.ring[i] < r.ring[j] })
}

// Del 删除哈希环中若干指定节点
func (r *HashRing) Del(nodes ...string) {
	// 加写锁
	r.lck.Lock()
	defer r.lck.Unlock()

	for _, node := range nodes {
		// 如果节点不存在,则不进行操作
		if _, ok := r.nodes[node]; !ok {
			r.logger.Info("node [%s] unexists", node)
			continue
		}
		// 删除对应节点的所有虚拟节点
		for i := 0; i < r.replicas; i++ {
			hash := r.hash([]byte(fmt.Sprintf("%d%s", i, node)))
			// 遍历所有节点
			for j := 0; j < len(r.ring); j++ {
				if hash == r.ring[j] {
					r.ring = append(r.ring[:j], r.ring[j+1:]...)
					break
				}
			}
			r.logger.Debug("delete vitural node [%d%s] - hash [%d]", i, node, hash)
			delete(r.hash2node, hash)
		}
		delete(r.nodes, node)
		r.logger.Info("node [%s] delete successfully", node)
	}
}

// Get 获取哈希环中key的哈希值最近的节点
func (r *HashRing) Get(key string) string {
	if r.IsEmpty() {
		return ""
	}

	// 加读锁
	r.lck.RLock()
	defer r.lck.RUnlock()

	hash := r.hash([]byte(key))
	r.logger.Debug("hash(%s) = %d", key, hash)

	// 在已排序的r.ring中进行二分搜索
	idx := sort.Search(len(r.ring), func(i int) bool { return r.ring[i] >= hash })
	r.logger.Info("get key [%s] from node [%s]", key, r.hash2node[r.ring[idx%len(r.ring)]])

	return r.hash2node[r.ring[idx%len(r.ring)]]
}
