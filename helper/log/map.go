package log

import (
	"sort"
	"sync"
)

// Map 代表以目录路径区分的日志的容器。
type Map interface {
	// Append 会对指定的目录路径的日志列表中追加一条日志。
	Append(key string, one One) (newKey, ok bool)
	// Get 会查找并返回指定的目录路径的所有日志。
	Get(key string) (list List, ok bool)
	// Delete 会删除指定目录路径及对应的日志。
	Delete(key string)
	// Range 会遍历目录路径及对应的日志并调用参数 f。
	// 若在遍历时 f 返回了 false 则会停止遍历。
	Range(f func(key string, list List) bool)
}

type sortedMap struct {
	sm sync.Map
}

// NewMap 会创建一个日志字典实例。
func NewMap() Map {
	return &sortedMap{}
}

func (m *sortedMap) Append(key string, one One) (newKey, ok bool) {
	if key == "" || one == nil {
		return
	}
	if value, ok := m.sm.Load(key); ok {
		list := value.(List)
		list.Append(one)
	} else {
		newKey = true
		list := NewList()
		list.Append(one)
		m.sm.Store(key, list)
	}
	ok = true
	return
}

func (m *sortedMap) Get(key string) (list List, ok bool) {
	if key == "" {
		return nil, false
	}
	value, ok := m.sm.Load(key)
	if ok && value != nil {
		list = value.(List)
	}
	return
}

func (m *sortedMap) Delete(key string) {
	if key == "" {
		return
	}
	m.sm.Delete(key)
}

func (m *sortedMap) Range(f func(key string, list List) bool) {
	keys := make([]string, 0)
	listMap := make(map[string]List)
	m.sm.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		keys = append(keys, keyStr)
		listMap[keyStr] = value.(List)
		return true
	})
	sort.Strings(keys)
	for _, key := range keys {
		if !f(key, listMap[key]) {
			break
		}
	}
}
