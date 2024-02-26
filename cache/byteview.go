package mycache

type ByteView struct {
	data []byte
}

// Len 返回ByteView实例的字节长度
func (v ByteView) Len() int {
	return len(v.data)
}

// ByteSlice 返回ByteView实例的数据拷贝
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.data)
}

// String 返回ByteView实例的字符串
func (v ByteView) String() string {
	return string(v.data)
}

func cloneBytes(data []byte) []byte {
	c := make([]byte, len(data))
	copy(c, data)
	return c
}
