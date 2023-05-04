package lru

import "container/list"

type Cache struct {
	maxBytes  int64                         // 最大内存，当maxBytes=0时，表示当前没有内存限制
	nBytes    int64                         // 当前使用的内存数
	ll        *list.List                    // lru的双链表
	cache     map[string]*list.Element      // string -> node
	OnEvicted func(key string, value Value) // 移除双链表中的node时触发的回调函数
}
type node struct { // list中的node的类型
	key   string // 在node保存key，可以反向映射删除map
	value Value
}

type Value interface { //
	Len() int // 返回占用的内存大小
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(s string) (value Value, ok bool) { // 通过map查找当前list是否存在key为s的结点
	if ele, ok := c.cache[s]; ok { // 当前结点存在
		c.ll.MoveToFront(ele)   // 将当前结点移动双链表头，从尾开始删除
		kv := ele.Value.(*node) // 这处是指将ele中的Value成员强转成*node类型，指针可以直接修改
		return kv.value, true   // 取出对应值返回
	}
	return
}

func (c *Cache) RemoveOldest() { // 移出最少访问的结点
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele) // 删除链表中的结点
		kv := ele.Value.(*node)
		delete(c.cache, kv.key)                                // 删除对应map的key
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len()) // 更新cache当前占用的内存
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value) // 回调函数
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { // 当前key存在，修改value即可
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*node)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else { // 当前key不存在，插入
		ele := c.ll.PushFront(&node{key, value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes { // 当前占用内存数超出最大内存，lru删除
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int { // 返回元素个数
	return c.ll.Len()
}
