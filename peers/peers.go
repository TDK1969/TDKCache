package peers

// PeerPicker接口根据传入的key选择相应的节点PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerServer interface {
	Set(peers ...string)
	PickPeer(key string) (PeerGetter, bool)
	Start(addrs []string, g GroupCache)
}

type GroupCache interface {
	RegisterPeers(peers PeerPicker)
}

// PeerGetter接口需要实现Get方法，从其他节点获取指定key的值

type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
