// Package singleflight 防止缓存击穿，限制请求
package singleflight

import "sync"

type call struct { // 代表正在进行 或 已经结束的请求
	wg  sync.WaitGroup // 加锁
	val interface{}    // 请求的值
	err error          // 错误
}

type Group struct { // 管理不同key的请求
	mu sync.Mutex
	m  map[string]*call // key->call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) { // 相当于包括了一下fn，每次根据请求key重新判断之后的请求
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call) // 实现懒加载
	}
	if c, ok := g.m[key]; ok { // 如果当前请求已经进行，则等待，反之则继续请求
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)  // 第一个请求
	g.m[key] = c // 加入map中
	g.mu.Unlock()

	c.val, c.err = fn() // fn执行完毕，值存入至call中
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key) // 将对应key的call删除，因为此次请求已经结束了
	g.mu.Unlock()

	return c.val, c.err
}
