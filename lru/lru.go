package lru

import (
	"container/list"
)

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

type Cache struct {
	size     int64
	cap      int64
	list     *list.List
	cache    map[string]*list.Element
	CallBack func(key , value string)
}

type entry struct {
	key   string
	value Value
}

func NewCache(cap int64, rollBack func(key, value string)) *Cache {
	return &Cache{
		cap:      cap,
		list:     list.New(),
		cache:    make(map[string]*list.Element),
		CallBack: rollBack,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if elem, ok := c.cache[key]; ok {
		c.list.MoveToBack(elem)
		return elem.Value.(*entry).value, ok
	}
	return
}

func (c *Cache) Update(key string, value Value) {
	if elem, ok := c.cache[key]; ok {
		c.list.MoveToBack(elem)
		kv := elem.Value.(*entry)
		c.size += int64(value.Len()) - int64(kv.value.Len())
		c.Clear()
		kv.value = value
	} else {
		c.size += int64(value.Len()) + int64(len(key))
		c.Clear()
		elem = c.list.PushBack(&entry{key: key, value: value})
		c.cache[key] = elem
	}
}

func (c *Cache) Clear() {
	for c.cap != 0 && c.size > c.cap {
		c.RemoveOldest()
	}
}

func (c *Cache) RemoveOldest() {
	elem := c.list.Front()
	if elem != nil {
		c.list.Remove(elem)
		kv := elem.Value.(*entry)
		delete(c.cache, kv.key)
		c.size -= int64(len(kv.key)) + int64(kv.value.Len())
	}

}

func (c *Cache) Len() int {
	return c.list.Len()
}
