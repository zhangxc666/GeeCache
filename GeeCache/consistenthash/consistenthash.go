package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32 // 哈希函数

type Map struct {
	hash     Hash           // 哈希函数
	replicas int            // 虚节点倍数
	keys     []int          // 哈希环
	hashMap  map[int]string // 映射虚拟节点-》真实节点  key是哈希值，值是真实节点
}

func New(replicas int, fn Hash) *Map { // 初始化一个一致性哈希的数据结构
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ { // 对每一个真实节点，创建多个虚拟节点
			hash := int(m.hash([]byte(strconv.Itoa(i) + key))) // 计算每个虚拟节点哈希值
			m.keys = append(m.keys, hash)                      // 虚拟节点加入至哈希环中
			m.hashMap[hash] = key                              // 虚拟节点找真实节点
		}
	}
	sort.Ints(m.keys) // 哈希环排序
}

func (m *Map) Get(keys string) string { // 找到距离当前key最近的一个存储结点
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(keys)))                  // 当前key的哈希值
	idx := sort.Search(len(m.keys), func(i int) bool { // 二分查第一个哈希值大于它虚拟节点的下标
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[idx%len(m.keys)]] // 虚拟节点反向映射至真实节点
}
