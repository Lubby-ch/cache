package cache

import (
	"catch/lru"
	"fmt"
	"log"
	"sync"
)

type NodeGetter interface {
	Get(group string, key string) ([]byte, error)
}

type IRemoteGetter interface {
	PickNode(key string) (NodeGetter, bool)
}

type ILocalGetter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type cache struct {
	mu  sync.RWMutex
	lru *lru.Cache
	cap int64
}

func NewCache(cap int64) *cache {
	return &cache{
		cap: cap,
		lru: lru.NewCache(cap, nil),
	}
}

func (c *cache) add(key string, value ByteValue) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = lru.NewCache(c.cap, nil)
	}

	c.lru.Update(key, value)
}

func (c *cache) get(key string) (value ByteValue, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	lv, ok := c.lru.Get(key)
	if ok {
		return *lv.(*ByteValue), ok
	}
	return
}

type Group struct {
	name         string
	cache        *cache
	localGetter  ILocalGetter  // 获取本地数据
	remoteGetter IRemoteGetter // 获取远程数据
	callMgr      *CallMgr
}

var (
	mu     sync.RWMutex
	groups map[string]*Group
)

func NewGroup(name string, cap int64, localGetter ILocalGetter, remoterGetter IRemoteGetter) *Group {
	if localGetter == nil {
		panic("ILocalGetter is nil")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:         name,
		cache:        NewCache(cap),
		localGetter:  localGetter,
		remoteGetter: remoterGetter,
		callMgr:      NewCallMgr(),
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (g *Group) Get(key string) (ByteValue, error) {
	if key == "" {
		return ByteValue{}, fmt.Errorf("key is nil")
	}
	if v, ok := g.cache.get(key); ok {
		log.Println("[Cache] hit", key)
		return v, nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (ByteValue, error) {
	val, err := g.callMgr.Do(key, func() (interface{}, error) {
		if g.remoteGetter != nil {
			if getter, ok := g.remoteGetter.PickNode(key); ok && getter != nil {
				value, err := g.getFromRemote(getter, key)
				if err == nil {
					return value, nil
				}
			}
		}
		return g.getFromLocal(key)
	})
	if err != nil {
		return ByteValue{}, err
	}
	return val.(ByteValue), nil
}

func (g *Group) getFromLocal(key string) (ByteValue, error) {
	bytes, err := g.localGetter.Get(key)
	if err != nil {
		return ByteValue{}, err
	}
	value := ByteValue{
		bytes: cloneBytes(bytes),
	}
	g.updateCache(key, value)
	return value, nil
}

func (g *Group) getFromRemote(node NodeGetter, key string) (ByteValue, error) {
	bytes, err := node.Get(g.name, key)
	if err != nil {
		return ByteValue{}, err
	}
	return ByteValue{bytes: bytes}, nil
}

func (g *Group) updateCache(key string, value ByteValue) {
	g.cache.add(key, value)
}
