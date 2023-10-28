package mysyncm

import (
	"fmt"
	"sync"
)

type SyncMap struct {
	mtx  sync.Mutex
	data map[int]int
}

func Empty() *SyncMap {
	return &SyncMap{data: make(map[int]int), mtx: sync.Mutex{}}
}

func New(data map[int]int) *SyncMap {
	return &SyncMap{data: data, mtx: sync.Mutex{}}
}

func (m *SyncMap) lock() {
	m.mtx.Lock()
}

func (m *SyncMap) unlock() {
	m.mtx.Unlock()
}

func (m *SyncMap) Set(key int, value int) *SyncMap {
	m.lock()
	defer m.unlock()

	m.data[key] = value
	return m
}

func (m *SyncMap) Get(key int) (int, bool) {
	m.lock()
	defer m.unlock()

	v, ok := m.data[key]
	return v, ok
}

func (m *SyncMap) Exists(key int) bool {
	m.lock()
	defer m.unlock()

	_, ok := m.data[key]
	return ok
}

func (m *SyncMap) Delete(key int) {
	delete(m.data, key)
}

func (m *SyncMap) GetData() map[int]int {
	m.lock()
	defer m.unlock()

	dataCopy := make(map[int]int)
	for k, v := range m.data {
		dataCopy[k] = v
	}
	return dataCopy
}

func (m *SyncMap) String() string {
	m.lock()
	defer m.unlock()

	return fmt.Sprint(m.data)
}
