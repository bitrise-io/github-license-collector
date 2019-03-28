package utils

import "sync"

type ThreadSafeCounter struct {
	Counter int
	mutex   sync.Mutex
}

func NewThreadSafeCounter() *ThreadSafeCounter {
	return &ThreadSafeCounter{}
}

func (tsc *ThreadSafeCounter) Increment() int {
	tsc.mutex.Lock()
	tsc.Counter++
	val := tsc.Counter
	tsc.mutex.Unlock()
	return val
}

func (tsc *ThreadSafeCounter) Value() int {
	tsc.mutex.Lock()
	val := tsc.Counter
	tsc.mutex.Unlock()
	return val
}
