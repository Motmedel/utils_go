package cache_group

import (
	"sync"
)

type storedCall[T any] struct {
	waitGroup sync.WaitGroup

	value T
	err   error
}

type Group[T any] struct {
	mutex sync.Mutex
	cache map[string]*storedCall[T]
}

func (group *Group[T]) Do(key string, fn func() (T, error)) (T, error) {
	group.mutex.Lock()

	if group.cache == nil {
		group.cache = make(map[string]*storedCall[T])
	}

	if call, ok := group.cache[key]; ok {
		group.mutex.Unlock()
		call.waitGroup.Wait()

		return call.value, call.err
	}

	call := new(storedCall[T])
	call.waitGroup.Add(1)
	group.cache[key] = call
	group.mutex.Unlock()

	group.performCall(call, fn)

	return call.value, call.err
}

func (group *Group[T]) performCall(call *storedCall[T], fn func() (T, error)) {
	defer call.waitGroup.Done()
	call.value, call.err = fn()
}
