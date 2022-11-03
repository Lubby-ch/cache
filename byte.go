package cache

type ByteValue struct {
	bytes []byte
}

func (v ByteValue) Len() int {
	return len(v.bytes)
}

func (v ByteValue) String() string {
	return string(v.bytes)
}

