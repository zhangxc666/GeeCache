package GeeCache

type ByteView struct { // 创建一个只读的数据结构表示缓存值
	b []byte
}

func (v ByteView) Len() int { // 返回占用字节数
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte { // 返回一个切片拷贝
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
