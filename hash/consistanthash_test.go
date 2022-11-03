package hash

import (
	"strconv"
	"testing"
)

func TestMap_Add(t *testing.T) {
	type args struct {
		keys []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "哈希一致性测试",
			args: args{
				keys: []string{
					"6",
					"4",
					"2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(3, nil)
			m.Add(tt.args.keys...)
		})
	}
}

func TestMap_Get(t *testing.T) {
	type args struct {
		keys []string
	}
	tests := []struct {
		name    string
		args    args
		getCase map[string]string
	}{
		{
			name: "哈希一致性测试",
			args: args{
				keys: []string{
					"6",
					"4",
					"2",
				},
			},
			getCase: map[string]string{
				"2":  "2",
				"11": "2",
				"23": "4",
				"27": "2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := New(3, func(key []byte) uint32 {
				i, _ := strconv.Atoi(string(key))
				return uint32(i)
			})
			hash.Add(tt.args.keys...)
			for k, v := range tt.getCase {
				if got := hash.Get(k); got != v {
					t.Errorf("Get() = %v, want %v", got, v)
				}
			}
		})
	}
}
