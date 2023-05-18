package GeeCache

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) // 传入key选择对应的节点PeerGetter
}
type PeerGetter interface { // 从group中查找key的缓存值
	Get(group string, key string) ([]byte, error)
}
