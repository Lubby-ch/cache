package cache

import "sync"

type session struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type CallMgr struct {
	mu     sync.Mutex
	record map[string]*session
}

func NewCallMgr() *CallMgr {
	return &CallMgr{
		record: make(map[string]*session),
	}
}

func (c *CallMgr) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	c.mu.Lock()
	if s, ok := c.record[key]; ok {
		c.mu.Unlock()
		s.wg.Wait()
		return s.val, s.err
	}
	s := new(session)
	s.wg.Add(1)
	c.mu.Unlock()
	s.val, s.err = fn()
	s.wg.Done()

	c.mu.Lock()
	delete(c.record, key)
	c.mu.Unlock()

	s.wg.Done()
	return s.val, s.err
}
