package peers

// PeerPicker接口根据传入的key选择相应的节点PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter接口需要实现Get方法，从其他节点获取指定key的值
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
