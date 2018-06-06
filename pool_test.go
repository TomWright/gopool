package gopool_test

import (
	"testing"
	"github.com/tomwright/gopool"
	"context"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestPool_StartStop(t *testing.T) {
	a := assert.New(t)

	startedWorkers := safeCounter{}
	finishedWorkers := safeCounter{}

	var work gopool.WorkFunc = func(ctx context.Context) error {
		startedWorkers.Inc()
		for {
			select {
			case <-ctx.Done():
				finishedWorkers.Inc()
				return ctx.Err()
			}
		}
	}

	var workerCount gopool.WorkerCountFunc = func() uint64 {
		return 3
	}

	var sleepTime gopool.SleepTimeFunc = func() time.Duration {
		return time.Millisecond * 500
	}

	p := gopool.NewPool("test", work, workerCount, sleepTime, context.Background())
	a.Equal(0, startedWorkers.Val())
	a.Equal(0, finishedWorkers.Val())

	cancel := p.Start()
	time.Sleep(time.Millisecond * 1000)
	a.Equal(3, startedWorkers.Val())
	a.Equal(0, finishedWorkers.Val())

	cancel()
	time.Sleep(time.Millisecond * 1000)
	a.Equal(3, startedWorkers.Val())
	a.Equal(3, finishedWorkers.Val())
}

func TestPool_DesiredWorkerCount_Increment(t *testing.T) {
	a := assert.New(t)

	wc := safeCounter{}
	wc.Set(3)
	pWc := &wc
	startedWorkers := safeCounter{}
	finishedWorkers := safeCounter{}

	var work gopool.WorkFunc = func(ctx context.Context) error {
		startedWorkers.Inc()
		for {
			select {
			case <-ctx.Done():
				finishedWorkers.Inc()
				return ctx.Err()
			}
		}
	}

	var workerCount gopool.WorkerCountFunc = func() uint64 {
		return uint64(pWc.Val())
	}

	var sleepTime gopool.SleepTimeFunc = func() time.Duration {
		return time.Millisecond * 500
	}

	p := gopool.NewPool("test", work, workerCount, sleepTime, context.Background())
	a.Equal(0, startedWorkers.Val())
	a.Equal(0, finishedWorkers.Val())

	cancel := p.Start()
	time.Sleep(time.Millisecond * 1000)
	a.Equal(3, startedWorkers.Val())
	a.Equal(0, finishedWorkers.Val())

	wc.Set(5)
	time.Sleep(time.Millisecond * 1000)
	a.Equal(5, startedWorkers.Val())
	a.Equal(0, finishedWorkers.Val())

	cancel()
	time.Sleep(time.Millisecond * 1000)
	a.Equal(5, startedWorkers.Val())
	a.Equal(5, finishedWorkers.Val())
}

func TestPool_DesiredWorkerCount_Decrement(t *testing.T) {
	a := assert.New(t)

	wc := safeCounter{}
	wc.Set(3)
	pWc := &wc
	startedWorkers := safeCounter{}
	finishedWorkers := safeCounter{}

	var work gopool.WorkFunc = func(ctx context.Context) error {
		startedWorkers.Inc()
		for {
			select {
			case <-ctx.Done():
				finishedWorkers.Inc()
				return ctx.Err()
			}
		}
	}

	var workerCount gopool.WorkerCountFunc = func() uint64 {
		return uint64(pWc.Val())
	}

	var sleepTime gopool.SleepTimeFunc = func() time.Duration {
		return time.Millisecond * 500
	}

	p := gopool.NewPool("test", work, workerCount, sleepTime, context.Background())
	a.Equal(0, startedWorkers.Val())
	a.Equal(0, finishedWorkers.Val())

	cancel := p.Start()
	time.Sleep(time.Millisecond * 1000)
	a.Equal(3, startedWorkers.Val())
	a.Equal(0, finishedWorkers.Val())

	wc.Set(1)
	time.Sleep(time.Millisecond * 1000)
	a.Equal(3, startedWorkers.Val())
	a.Equal(2, finishedWorkers.Val())

	cancel()
	time.Sleep(time.Millisecond * 1000)
	a.Equal(3, startedWorkers.Val())
	a.Equal(3, finishedWorkers.Val())
}
