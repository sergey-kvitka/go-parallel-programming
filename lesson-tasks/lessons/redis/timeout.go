package redis

import (
	"sync"
	"time"
)

type timeoutInfo[K KeyType] struct {
	mtx  *sync.Mutex
	data map[K]time.Time
}

func newTimeout[K KeyType]() timeoutInfo[K] {
	return timeoutInfo[K]{
		mtx:  &sync.Mutex{},
		data: make(map[K]time.Time),
	}
}
