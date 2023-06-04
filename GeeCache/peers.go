package GeeCache

import pb "GeeCache/GeeCache/geecachepb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) // 传入key选择对应的节点PeerGetter
}
type PeerGetter interface { // 从group中查找key的缓存值
	Get(in *pb.Request, out *pb.Response) error
}
