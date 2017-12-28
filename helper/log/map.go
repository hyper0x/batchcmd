package log

import (
	"sort"
	"sync"
)

// Map represents a container for logs that
// are distinguished by directory paths.
type Map interface {
	// Append appends a log to the log list of
	// the specified directory path.
	Append(key string, one One) (newKey, ok bool)
	// Get finds and returns all logs of the specified
	// directory path.
	Get(key string) (list List, ok bool)
	// Delete deletes the specified directory path and
	// its corresponding logs.
	Delete(key string)
	// The Range traverses the directory path and the corresponding log,
	// and call the parameter f.
	// If f returns false, then it will stop traversing.
	Range(f func(key string, list List) bool)
}

type sortedMap struct {
	sm sync.Map
}

// NewMap creates a log dictionary instance.
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
