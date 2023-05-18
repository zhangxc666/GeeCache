package GeeCache

import (
	"fmt"
	"log"
	"sync"
)

type Getter interface { // 回调函数的接口 当缓存不存在时，调用这个函数，得到源数据，由用户设计这个函数
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error) // 接口型函数

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct { // 一个缓存命名空间
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
}

var (
	mutex  sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group { // 创建一个组实例
	if getter == nil {
		panic("nil Getter")
	}
	mutex.Lock() // 加锁处理
	defer mutex.Unlock()
	g := &Group{ // 创建一个新的组
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g // 加到组中，用map[group_name]->group找到组
	return g
}

func GetGroup(name string) *Group { // 通过name返回当前group
	mutex.RLock()
	defer mutex.RUnlock()
	g := groups[name]
	return g
}

func (g *Group) RegisterPeers(peer PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peer
}

func (g *Group) Get(key string) (ByteView, error) { // 查找缓存值
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok { // 缓存值在当前主cache存在，则返回
		log.Println("[GeeCache]hit")
		return v, nil
	}
	return g.load(key) // 当前主cache缓存值不存在调用load
}

func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}
	return g.getLocally(key) //添加至自己当前的main缓存中
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key) // 通过用户定义回调函数获取数据
	if err != nil {                 // 当前数据传入出错
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)} // 获取了value
	g.populateCache(key, value)             // 加到当前的maincache之中
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
