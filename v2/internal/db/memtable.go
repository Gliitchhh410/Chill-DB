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

// Put adds a key-value pair to the table.
// TODO:
// 1. Lock the mutex (m.mu.Lock())
// 2. Defer the unlock
// 3. Add the key and value to m.data
// 4. Update m.size (add len(key) + len(value))
func (m *MemTable) Put(key string, value []byte){
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
	m.size += len(key) + len(value)
}


// Get retrieves a value by key.
// TODO:
// 1. Read-Lock the mutex (m.mu.RLock()) - allows multiple readers, blocks writers
// 2. Defer the unlock
// 3. Look up the key in m.data
// 4. Return the value and the exists boolean
func (m *MemTable) Get(key string) ([]byte, bool){
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, exists := m.data[key]
	if exists {
		return value, true
	}
	return nil, false
}





