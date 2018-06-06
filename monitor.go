package gopool

import (
	"time"
	"fmt"
)

// SleepTimeFunc should return a time.Duration that we should sleep between checking the worker count
// of a pool
type SleepTimeFunc func() time.Duration

func monitorPool(pool *Pool) {
	for {
		select {
		case <-pool.Context().Done():
			return
		default:
			pool.mu.Lock()

			diff := int(pool.desiredWorkerCount()) - len(pool.workers)
			if diff < 0 {
				removeWorkersFromPool(pool, uint64(-diff))
			} else if diff > 0 {
				addWorkersToPool(pool, uint64(diff))
			}

			sleepTime := pool.sleepTime()

			pool.mu.Unlock()

			time.Sleep(sleepTime)
		}
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
		pool.workers[x].cancel()
	}
}

func removeWorkerFromPoolOnceDone(pool *Pool, worker poolWorker) {
	<-worker.worker.Done()

	pool.mu.Lock()
	subjectWorkerId := worker.worker.ID()
	for k, w := range pool.workers {
		if subjectWorkerId == w.worker.ID() {
			pool.workers = append(pool.workers[:k], pool.workers[k+1:]...)
		}
	}
	pool.mu.Unlock()
}
