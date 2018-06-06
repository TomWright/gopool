package gopool

import (
	"sync"
	"context"
)

// WorkerCountFunc should return the number of workers you want to be running in the pool
type WorkerCountFunc func() uint64

// NewPool returns a new pool
func NewPool(id string, work WorkFunc, desiredWorkerCount WorkerCountFunc, sleepTime SleepTimeFunc, ctx context.Context) *Pool {
	p := new(Pool)
	p.id = id
	p.work = work
	p.workers = make([]poolWorker, 0)
	p.desiredWorkerCount = desiredWorkerCount
	p.sleepTime = sleepTime
	p.ctx = ctx
	p.ctx = context.WithValue(p.ctx, "poolId", p.id)
	return p
}

type poolWorker struct {
	cancel context.CancelFunc
	worker *Worker
}

// Pool represents a group of workers performing the same task simultaneously
type Pool struct {
	mu                 sync.Mutex
	id                 string
	work               WorkFunc
	workers            []poolWorker
	desiredWorkerCount WorkerCountFunc
	sleepTime          SleepTimeFunc
	ctx                context.Context
	cancel             context.CancelFunc
}

// ID returns the pools unique ID
func (p *Pool) ID() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.id
}

// Context
func (p *Pool) Context() context.Context {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.ctx
}

// Start initiates the pool monitor
func (p *Pool) Start() context.CancelFunc {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ctx, p.cancel = context.WithCancel(p.ctx)

	go monitorPool(p)

	return p.cancel
}
