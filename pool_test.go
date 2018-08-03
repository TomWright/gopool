package gopool_test

import (
	"testing"
	. "github.com/tomwright/gopool"
	"context"
	"time"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestPool_ID(t *testing.T) {
	t.Parallel()

	a := assert.New(t)

	var work WorkFunc
	var workerCount WorkerCountFunc
	var sleepTime SleepTimeFunc

	p := NewPool("some-id", work, workerCount, sleepTime, context.Background())

	a.Equal("some-id", p.ID())
}

func TestPool_StartStop(t *testing.T) {
	t.Parallel()

	startedWorkers := safeCounter{}
	finishedWorkers := safeCounter{}

	var work WorkFunc = func(ctx context.Context) error {
		startedWorkers.Inc()
		for {
			select {
			case <-ctx.Done():
				finishedWorkers.Inc()
				return ctx.Err()
			}
		}
	}

	var workerCount WorkerCountFunc = func() uint64 {
		return 3
	}

	var sleepTime SleepTimeFunc = func() time.Duration {
		return time.Millisecond * 100
	}

	safeCounterHasVal := func(started *safeCounter, startedVal int, finished *safeCounter, finishedVal int) func() error {
		return func() error {
			got := started.Val()
			if got != startedVal {
				return fmt.Errorf("expected %d workers to be started, got %d", startedVal, got)
			}
			got = finished.Val()
			if got != finishedVal {
				return fmt.Errorf("expected %d workers to be finished, got %d", finishedVal, got)
			}
			return nil
		}
	}

	p := NewPool("test", work, workerCount, sleepTime, context.Background())
	if err := timeoutAfter(time.Second, safeCounterHasVal(&startedWorkers, 0, &finishedWorkers, 0)); err != nil {
		t.Fatal(err)
	}

	cancel := p.Start()
	if err := timeoutAfter(time.Second, safeCounterHasVal(&startedWorkers, 3, &finishedWorkers, 0)); err != nil {
		t.Fatal(err)
	}

	cancel()
	if err := timeoutAfter(time.Second, safeCounterHasVal(&startedWorkers, 3, &finishedWorkers, 3)); err != nil {
		t.Fatal(err)
	}
}

func TestPool_DesiredWorkerCount_Increment(t *testing.T) {
	t.Parallel()

	wc := safeCounter{}
	wc.Set(3)
	pWc := &wc
	startedWorkers := safeCounter{}
	finishedWorkers := safeCounter{}

	var work WorkFunc = func(ctx context.Context) error {
		startedWorkers.Inc()
		for {
			select {
			case <-ctx.Done():
				finishedWorkers.Inc()
				return ctx.Err()
			}
		}
	}

	safeCounterHasVal := func(started *safeCounter, startedVal int, finished *safeCounter, finishedVal int) func() error {
		return func() error {
			got := started.Val()
			if got != startedVal {
				return fmt.Errorf("expected %d workers to be started, got %d", startedVal, got)
			}
			got = finished.Val()
			if got != finishedVal {
				return fmt.Errorf("expected %d workers to be finished, got %d", finishedVal, got)
			}
			return nil
		}
	}

	var workerCount WorkerCountFunc = func() uint64 {
		return uint64(pWc.Val())
	}

	var sleepTime SleepTimeFunc = func() time.Duration {
		return time.Millisecond * 100
	}

	p := NewPool("test", work, workerCount, sleepTime, context.Background())
	if err := timeoutAfter(time.Second, safeCounterHasVal(&startedWorkers, 0, &finishedWorkers, 0)); err != nil {
		t.Fatal(err)
	}

	cancel := p.Start()
	if err := timeoutAfter(time.Second, safeCounterHasVal(&startedWorkers, 3, &finishedWorkers, 0)); err != nil {
		t.Fatal(err)
	}

	wc.Set(3)
	if err := timeoutAfter(time.Second, safeCounterHasVal(&startedWorkers, 3, &finishedWorkers, 0)); err != nil {
		t.Fatal(err)
	}

	wc.Set(5)
	if err := timeoutAfter(time.Second, safeCounterHasVal(&startedWorkers, 5, &finishedWorkers, 0)); err != nil {
		t.Fatal(err)
	}

	wc.Set(6)
	if err := timeoutAfter(time.Second, safeCounterHasVal(&startedWorkers, 6, &finishedWorkers, 0)); err != nil {
		t.Fatal(err)
	}

	cancel()
	if err := timeoutAfter(time.Second, safeCounterHasVal(&startedWorkers, 6, &finishedWorkers, 6)); err != nil {
		t.Fatal(err)
	}
}

func TestPool_DesiredWorkerCount_Increment_Decrement(t *testing.T) {
	t.Parallel()

	wc := safeCounter{}
	wc.Set(3)
	pWc := &wc
	startedWorkers := &safeCounter{}
	finishedWorkers := &safeCounter{}

	var work WorkFunc = func(ctx context.Context) error {
		startedWorkers.Inc()
		for {
			select {
			case <-ctx.Done():
				finishedWorkers.Inc()
				return ctx.Err()
			}
		}
	}

	var workerCount WorkerCountFunc = func() uint64 {
		return uint64(pWc.Val())
	}

	var sleepTime SleepTimeFunc = func() time.Duration {
		return time.Millisecond * 100
	}

	safeCounterHasVal := func(started *safeCounter, startedVal int, finished *safeCounter, finishedVal int) func() error {
		return func() error {
			got := started.Val()
			if got != startedVal {
				return fmt.Errorf("expected %d workers to be started, got %d", startedVal, got)
			}
			got = finished.Val()
			if got != finishedVal {
				return fmt.Errorf("expected %d workers to be finished, got %d", finishedVal, got)
			}
			return nil
		}
	}

	p := NewPool("test", work, workerCount, sleepTime, context.Background())
	if err := timeoutAfter(time.Second, safeCounterHasVal(startedWorkers, 0, finishedWorkers, 0)); err != nil {
		t.Fatal(err)
	}

	cancel := p.Start()
	if err := timeoutAfter(time.Second, safeCounterHasVal(startedWorkers, 3, finishedWorkers, 0)); err != nil {
		t.Fatal(err)
	}

	wc.Set(2)
	if err := timeoutAfter(time.Second, safeCounterHasVal(startedWorkers, 3, finishedWorkers, 1)); err != nil {
		t.Fatal(err)
	}

	wc.Set(5)
	if err := timeoutAfter(3*time.Second, safeCounterHasVal(startedWorkers, 6, finishedWorkers, 1)); err != nil {
		t.Fatal(err)
	}

	wc.Set(1)
	if err := timeoutAfter(3*time.Second, safeCounterHasVal(startedWorkers, 6, finishedWorkers, 5)); err != nil {
		t.Fatal(err)
	}

	wc.Set(0)
	if err := timeoutAfter(3*time.Second, safeCounterHasVal(startedWorkers, 6, finishedWorkers, 6)); err != nil {
		t.Fatal(err)
	}

	cancel()
	if err := timeoutAfter(time.Second, safeCounterHasVal(startedWorkers, 6, finishedWorkers, 6)); err != nil {
		t.Fatal(err)
	}
}
