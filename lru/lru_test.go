package lru

import (
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestCache_Get(t *testing.T) {
	keys := make([]string, 0)
	callBack := func(key string, value string) {
		keys = append(keys, key)
	}

	lru := NewCache(14, callBack)
	tests := []struct {
		key   string
		value string
	}{
		{
			key:   "key1",
			value: "value1",
		},
		{
			key:   "key2",
			value: "value2",
		},
		{
			key:   "k3",
			value: "v3",
		},
	}
	for _, test := range tests {
		lru.Update(test.key, String(test.value))
	}
	if _, ok := lru.Get("key2"); !ok  {
		t.Fatalf("Removeoldest key2 failed")
	}
}
