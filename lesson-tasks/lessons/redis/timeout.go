package redis

import (
	"sync"
	"time"
)

// timeoutInfo хранит информацию о ключах, данные по которым
// необходимо будет удалить, и время, когда это нужно будет сделать
type timeoutInfo[K KeyType] struct {
	mtx  *sync.Mutex
	data map[K]time.Time
}

// newTimeout создает экземпляр типа timeoutInfo
func newTimeout[K KeyType]() timeoutInfo[K] {
	return timeoutInfo[K]{
		mtx:  &sync.Mutex{},
		data: make(map[K]time.Time),
	}
}
