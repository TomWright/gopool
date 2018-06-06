package gopool

import (
	"context"
	"sync"
)

// WorkFunc defines a piece of work to be processed
type WorkFunc func(ctx context.Context) error

// NewWorker returns a worker instance
func NewWorker(id string, work WorkFunc, ctx context.Context) *Worker {
	w := new(Worker)
	w.id = id
	w.work = work
	w.ctx = ctx
	w.ctx = context.WithValue(w.ctx, "workerId", w.id)
	return w
}

// Worker is the definition for a single worker
type Worker struct {
	mu   sync.Mutex
	id   string
	work WorkFunc
	ctx  context.Context
	err  error
	done chan struct{}
}

// ID returns the works unique ID
func (w *Worker) ID() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.id
}

// Err returns an err if one occurred in the worker
func (w *Worker) Err() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.err
}

// Context returns the workers context
func (w *Worker) Context() context.Context {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ctx
}

// Start initiates a go routine for the worker and returns the cancel context
func (w *Worker) Start() context.CancelFunc {
	w.mu.Lock()
	w.done = make(chan struct{})
	w.mu.Unlock()
	ctx, cancel := context.WithCancel(w.Context())
	go func() {
		w.mu.Lock()
		w.err = w.work(ctx)
		close(w.done)
		w.mu.Unlock()
	}()
	return cancel
}

// Done returns a channel you can use to pick up on when a worker has finished
func (w *Worker) Done() chan struct{} {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.done
}
