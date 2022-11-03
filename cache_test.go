package cache

import (
	"reflect"
	"testing"
)

func TestGetter (t *testing.T) {
	var f ILocalGetter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}
