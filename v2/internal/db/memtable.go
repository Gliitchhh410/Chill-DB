package db


import (
	"sync"
)

type MemTable struct {
	data map[string][]byte
	mu sync.RWMutex
	size int
}


func NewMemTable() *MemTable {
	return &MemTable{
		data: make(map[string][]byte),
		size: 0,
	}
}


func (m *MemTable) Put(key string, value []byte){
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
	m.size += len(key) + len(value)
}



func (m *MemTable) Get(key string) ([]byte, bool){
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, exists := m.data[key]
	if exists {
		return value, true
	}
	return nil, false
}





