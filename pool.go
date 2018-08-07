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
	p.monitorPoolSize = true
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
	monitorPoolSize    bool
	running            bool
	doneChan           chan struct{}
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
func (p *Pool) Start() (context.CancelFunc, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return nil, ErrPoolAlreadyRunning
	}
	if p.sleepTime == nil {
		return nil, ErrPoolHasNilSleepTimeFunc
	}

	p.monitorPoolSize = true

	var cancel context.CancelFunc

	p.ctx, cancel = context.WithCancel(p.ctx)

	p.running = true
	p.doneChan = make(chan struct{})
	wrappedCancel := getWrappedContextCancelFunc(p, cancel)

	p.cancel = wrappedCancel

	go monitorPool(p)

	return wrappedCancel, nil
}

func getWrappedContextCancelFunc(p *Pool, cancel context.CancelFunc) context.CancelFunc {
	executedMu := &sync.Mutex{}
	executedBool := false
	var wrappedCancel context.CancelFunc = func() {
		executedMu.Lock()
		if ! executedBool {
			executedBool = true
			executedMu.Unlock()

			p.mu.Lock()
			p.running = false
			close(p.doneChan)
			p.mu.Unlock()
			cancel()
		} else {
			executedMu.Unlock()
		}
	}
	return wrappedCancel
}

func (p *Pool) StartOnce() (context.CancelFunc, error) {
	if p.Running() {
		return nil, ErrPoolAlreadyRunning
	}
	p.mu.Lock()
	p.monitorPoolSize = false

	var cancel context.CancelFunc

	p.ctx, cancel = context.WithCancel(p.ctx)

	p.running = true
	p.doneChan = make(chan struct{})
	wrappedCancel := getWrappedContextCancelFunc(p, cancel)

	p.cancel = wrappedCancel
	p.mu.Unlock()

	assertPoolSize(p)

	p.mu.Lock()
	// create a wait group which will close when all workers have finished
	wg := sync.WaitGroup{}
	workerCount := len(p.workers)
	wg.Add(workerCount)
	for _, w := range p.workers {
		internalW := w
		go func(w poolWorker, wg *sync.WaitGroup) {
			select {
			case <-w.worker.Done():
				wg.Done()
			}
		}(internalW, &wg)
	}
	p.mu.Unlock()

	// wait for all workers to finish and close the done chan
	go func(cancel context.CancelFunc, wg *sync.WaitGroup) {
		wg.Wait()
		cancel()
	}(p.cancel, &wg)

	return wrappedCancel, nil
}

// Running returns whether or not the pool is running
func (p *Pool) Running() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.running
}

// Done returns a channel that is closed when the pool is no longer running.
// Done returns nil if called when the pool hasn't started running yet.
func (p *Pool) Done() <-chan struct{} {
	p.mu.Lock()
	defer p.mu.Unlock()
	if ! p.running {
		return nil
	}
	return p.doneChan
}
