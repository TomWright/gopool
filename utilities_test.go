package gopool_test

import (
	"sync"
)

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

type safeIntSlice struct {
	mu    sync.Mutex
	slice []int
}

func (c *safeIntSlice) Init(val []int) *safeIntSlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.slice = val
	return c
}

func (c *safeIntSlice) Set(key int, val int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.slice[key] = val
}

func (c *safeIntSlice) Get(key int) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.slice[key]
}

func (c *safeIntSlice) Slice() []int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.slice
}
