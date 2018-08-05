package gopool_test

import (
	"testing"
	. "github.com/tomwright/gopool"
	"context"
	"time"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func ExamplePool() {
	startedWorkers := safeCounter{}
	finishedWorkers := safeCounter{}

	workInputChan := make(chan string, 5)

	var work WorkFunc = func(ctx context.Context) error {
		startedWorkers.Inc()
		fmt.Println("Starting worker")
		for {
			select {
			case <-ctx.Done():
				finishedWorkers.Inc()
				fmt.Println("Finishing worker")
				return ctx.Err()
			case x, ok := <-workInputChan:
				if ! ok {
					return fmt.Errorf("unexpected closed input chan")
				}
				fmt.Printf("Doing work: %s\n", x)
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
		panic(err)
	}
	fmt.Printf("%d started - %d finished\n", startedWorkers.Val(), finishedWorkers.Val())

	cancel := p.Start()
	defer cancel()
	if err := timeoutAfter(time.Second, safeCounterHasVal(&startedWorkers, 3, &finishedWorkers, 0)); err != nil {
		panic(err)
	}
	fmt.Printf("%d started - %d finished\n", startedWorkers.Val(), finishedWorkers.Val())

	// Note: sleeps are used to ensure output order
	workInputChan <- "One"
	time.Sleep(time.Millisecond * 20)
	workInputChan <- "Two"
	time.Sleep(time.Millisecond * 20)
	workInputChan <- "Three"

	time.Sleep(time.Millisecond * 20)

	cancel()
	if err := timeoutAfter(time.Second, safeCounterHasVal(&startedWorkers, 3, &finishedWorkers, 3)); err != nil {
		panic(err)
	}
	fmt.Printf("%d started - %d finished\n", startedWorkers.Val(), finishedWorkers.Val())

	// Output: 0 started - 0 finished
	// Starting worker
	// Starting worker
	// Starting worker
	// 3 started - 0 finished
	// Doing work: One
	// Doing work: Two
	// Doing work: Three
	// Finishing worker
	// Finishing worker
	// Finishing worker
	// 3 started - 3 finished
}

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
	defer cancel()
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
	defer cancel()
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
	defer cancel()
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
