package gopool_test

import "sync"

type safeCounter struct {
	mu      sync.Mutex
	counter int
}

func (c *safeCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counter++
}

func (c *safeCounter) Set(val int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counter = val
}

func (c *safeCounter) Val() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counter
}
