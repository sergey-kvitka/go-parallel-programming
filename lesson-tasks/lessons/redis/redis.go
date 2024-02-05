package redis

import (
	"sync"
	"time"
)

type KeyType interface {
	int | float64 | string
}

type ValueType interface {
	any
}

type Instance[K KeyType, V ValueType] struct {
	data map[K]V
	mtx  *sync.RWMutex

	timeouts timeoutInfo[K]
	interval time.Duration
}

func NewInstance[K KeyType, V ValueType](interval time.Duration) Instance[K, V] {
	return Instance[K, V]{
		data: make(map[K]V),
		mtx:  &sync.RWMutex{},

		timeouts: newTimeout[K](),
		interval: interval,
	}
}

func NewInstance1S[K KeyType, V ValueType]() Instance[K, V] {
	return NewInstance[K, V](time.Second)
}

func (r *Instance[K, V]) Get(key K) V {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.data[key]
}

func (r *Instance[K, V]) Set(key K, value V) {
	r.mtx.Lock()
	r.timeouts.mtx.Lock()

	defer r.mtx.Unlock()
	defer r.timeouts.mtx.Unlock()

	r.data[key] = value
	delete(r.timeouts.data, key)
}

func (r *Instance[K, V]) SetWithTimeout(key K, value V, expiration time.Duration) {
	r.mtx.Lock()
	r.timeouts.mtx.Lock()

	defer r.mtx.Unlock()
	defer r.timeouts.mtx.Unlock()

	r.data[key] = value
	r.timeouts.data[key] = time.Now().Add(expiration)
	if len(r.timeouts.data) == 1 {
		r.startTimeoutCheck()
	}
}

func (r *Instance[K, V]) startTimeoutCheck() {
	go func() {
		timeoutData := r.timeouts.data

		for now := range time.Tick(r.interval) {
			r.timeouts.mtx.Lock()

			for k, v := range timeoutData {
				if !now.Before(v) {
					r.mtx.Lock()
					delete(timeoutData, k)
					delete(r.data, k)
					r.mtx.Unlock()
				}
			}

			if len(timeoutData) == 0 {
				r.timeouts.mtx.Unlock()
				return
			}
			r.timeouts.mtx.Unlock()
		}
	}()
}
