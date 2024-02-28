package peers

import "TDKCache/peers/protobuf/pb"

// PeerPicker接口根据传入的key选择相应的节点PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter接口需要实现Get方法，从其他节点获取指定key的值
// HTTP
/*
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
*/

// ProtoBuf
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
