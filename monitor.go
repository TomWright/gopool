package gopool

import (
	"time"
	"fmt"
)

// SleepTimeFunc should return a time.Duration that we should sleep between checking the worker count
// of a pool
type SleepTimeFunc func() time.Duration

func monitorPool(pool *Pool) {
	ctx := pool.Context()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			assertPoolSize(pool)
			sleepTime := pool.sleepTime()

			time.Sleep(sleepTime)
		}
	}
}

func assertPoolSize(pool *Pool) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	diff := int(pool.desiredWorkerCount()) - len(pool.workers)
	if diff < 0 {
		removeWorkersFromPool(pool, uint64(-diff))
	} else if diff > 0 {
		addWorkersToPool(pool, uint64(diff))
	}
}

func addWorkersToPool(pool *Pool, add uint64) {
	for x := uint64(0); x < add; x++ {
		worker := poolWorker{
			worker: NewWorker(
				fmt.Sprintf("%s-%d", pool.id, time.Now().UnixNano()),
				pool.work,
				pool.ctx,
			),
		}
		worker.cancel = worker.worker.Start()
		pool.workers = append(pool.workers, worker)

		go removeWorkerFromPoolOnceDone(pool, worker)
	}
}

func removeWorkersFromPool(pool *Pool, remove uint64) {
	for x := uint64(0); x < remove; x++ {
		if x >= uint64(len(pool.workers)) {
			return
		}
		pool.workers[x].cancel()
	}
}

func removeWorkerFromPoolOnceDone(pool *Pool, worker poolWorker) {
	<-worker.worker.Done()

	pool.mu.Lock()
	defer pool.mu.Unlock()

	newWorkers := make([]poolWorker, 0)

	subjectWorkerId := worker.worker.ID()

	for _, w := range pool.workers {
		if subjectWorkerId != w.worker.ID() {
			newWorkers = append(newWorkers, w)
		}
	}

	pool.workers = newWorkers
}
