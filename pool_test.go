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

	cancel, err := p.Start()
	if err != nil {
		panic(err)
	}
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
	time.Sleep(time.Millisecond * 50)
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

	cancel, err := p.Start()
	assert.NoError(t, err)
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

	cancel, err := p.Start()
	assert.NoError(t, err)
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

	cancel, err := p.Start()
	assert.NoError(t, err)
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

// StartOnce will start a pool with the desired worker count, and will close the done
// channel when all workers have finished.
func ExamplePool_StartOnce() {
	ctx := context.TODO()

	// outputSlice is just a concurrently safe slice
	outputSlice := &safeIntSlice{}
	// notice that it initially contains 10 values of 0
	outputSlice.Init([]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

	workerInput := make(chan int)

	var work WorkFunc = func(ctx context.Context) error {
		fmt.Println("worker started")
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case key, ok := <-workerInput:
				if ! ok {
					return nil
				}
				// when we are given a key, the worker will set it to a value of 1
				outputSlice.Set(key, 1)
			}
		}
	}

	// in this case we will have 3 workers
	var workerCount WorkerCountFunc = func() uint64 {
		return 3
	}

	p := NewPool("test", work, workerCount, nil, ctx)

	fmt.Println("check #1")
	for i := 0; i < 10; i++ {
		fmt.Printf("%d = %d\n", i, outputSlice.Get(i))
	}

	fmt.Println("starting workers")

	cancel, err := p.StartOnce()
	if err != nil {
		panic(err)
	}
	defer cancel()
	<-time.After(time.Millisecond * 100)

	workerInput <- 0
	workerInput <- 2
	workerInput <- 4
	workerInput <- 6
	workerInput <- 8

	<-time.After(time.Millisecond * 100)
	fmt.Println("check #2")
	for i := 0; i < 10; i++ {
		fmt.Printf("%d = %d\n", i, outputSlice.Get(i))
	}

	workerInput <- 1
	workerInput <- 3
	workerInput <- 5
	workerInput <- 7

	fmt.Println("closing worker input chan")
	close(workerInput)

	select {
	case <-p.Done():
		fmt.Println("done")
	case <-time.After(time.Second):
		panic("timeout")
	}

	fmt.Println("check #3")
	for i := 0; i < 10; i++ {
		fmt.Printf("%d = %d\n", i, outputSlice.Get(i))
	}

	// Output: check #1
	// 0 = 0
	// 1 = 0
	// 2 = 0
	// 3 = 0
	// 4 = 0
	// 5 = 0
	// 6 = 0
	// 7 = 0
	// 8 = 0
	// 9 = 0
	// starting workers
	// worker started
	// worker started
	// worker started
	// check #2
	// 0 = 1
	// 1 = 0
	// 2 = 1
	// 3 = 0
	// 4 = 1
	// 5 = 0
	// 6 = 1
	// 7 = 0
	// 8 = 1
	// 9 = 0
	// closing worker input chan
	// done
	// check #3
	// 0 = 1
	// 1 = 1
	// 2 = 1
	// 3 = 1
	// 4 = 1
	// 5 = 1
	// 6 = 1
	// 7 = 1
	// 8 = 1
	// 9 = 0
}

func TestPool_StartOnce(t *testing.T) {
	t.Parallel()

	outputSlice := &safeIntSlice{}
	outputSlice.Init([]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	workerInput := make(chan int, 10)

	var work WorkFunc = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case key, ok := <-workerInput:
				if ! ok {
					return nil
				}
				outputSlice.Set(key, 1)
			}
		}
	}

	var workerCount WorkerCountFunc = func() uint64 {
		return 3
	}

	for i := 0; i < 10; i ++ {
		workerInput <- i
	}

	p := NewPool("test", work, workerCount, nil, context.Background())

	if expected, got := false, p.Running(); expected != got {
		t.Fatalf("expected pool.Running() to be %t, got %t", expected, got)
	}
	for i := 0; i < 10; i ++ {
		if expected, got := 0, outputSlice.Get(i); expected != got {
			t.Fatalf("expected value of %d, got %d", expected, got)
		}
	}

	cancel, err := p.StartOnce()
	assert.NoError(t, err)
	defer cancel()

	if expected, got := true, p.Running(); expected != got {
		t.Fatalf("expected pool.Running() to be %t, got %t", expected, got)
	}

	close(workerInput)

	select {
	case <-time.After(5 * time.Second):
		t.Fatalf("workers had not completed work after 5 seconds. output slice: %v", outputSlice.Slice())
	case <-p.Done():
	}

	if expected, got := false, p.Running(); expected != got {
		t.Fatalf("expected pool.Running() to be %t, got %t", expected, got)
	}

	for i := 0; i < 10; i ++ {
		if expected, got := 1, outputSlice.Get(i); expected != got {
			t.Fatalf("expected value of %d, got %d", expected, got)
		}
	}
}

func TestPool_Start_ErrPoolAlreadyRunning(t *testing.T) {
	t.Parallel()

	var work WorkFunc = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	var workerCount WorkerCountFunc = func() uint64 {
		return 1
	}

	var sleepTime SleepTimeFunc = func() time.Duration {
		return time.Millisecond * 100
	}

	p := NewPool("test", work, workerCount, sleepTime, context.Background())

	cancel, err := p.Start()
	assert.NoError(t, err)
	assert.NotNil(t, cancel)
	defer cancel()

	secondCancel, err := p.Start()
	assert.Error(t, err)
	assert.Nil(t, secondCancel)

	cancel()
}

func TestPool_Done_ReturnsCorrectValue(t *testing.T) {
	t.Parallel()

	var work WorkFunc = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	var workerCount WorkerCountFunc = func() uint64 {
		return 1
	}

	var sleepTime SleepTimeFunc = func() time.Duration {
		return time.Millisecond * 100
	}

	p := NewPool("test", work, workerCount, sleepTime, context.Background())

	var expected <-chan struct{}
	var unexpected <-chan struct{}
	var got <-chan struct{}
	if expected, got = nil, p.Done(); expected != got {
		t.Fatalf("expected p.Done() to be %v, got %v", expected, got)
	}

	cancel, err := p.Start()
	defer cancel()
	assert.NoError(t, err)
	assert.NotNil(t, cancel)

	if unexpected, got = nil, p.Done(); unexpected == got {
		t.Fatalf("did not expect p.Done() to be %v, got %v", unexpected, got)
	}

	cancel()

	if expected, got = nil, p.Done(); expected != got {
		t.Fatalf("expected p.Done() to be %v, got %v", expected, got)
	}
}

func TestPool_Done_ClosesCorrectly_PositiveCheck(t *testing.T) {
	t.Parallel()

	var work WorkFunc = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	var workerCount WorkerCountFunc = func() uint64 {
		return 1
	}

	var sleepTime SleepTimeFunc = func() time.Duration {
		return time.Millisecond * 100
	}

	p := NewPool("test", work, workerCount, sleepTime, context.Background())

	cancel, err := p.Start()
	defer cancel()
	assert.NoError(t, err)
	assert.NotNil(t, cancel)

	go func() {
		time.Sleep(time.Millisecond * 200)
		cancel()
	}()

	select {
	case <-p.Done():
	case <-time.After(time.Millisecond * 400):
		t.Fatalf("timed out before done channel was closed")
	}
}

func TestPool_Done_ClosesCorrectly_NegativeCheck(t *testing.T) {
	t.Parallel()

	var work WorkFunc = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	var workerCount WorkerCountFunc = func() uint64 {
		return 1
	}

	var sleepTime SleepTimeFunc = func() time.Duration {
		return time.Millisecond * 100
	}

	p := NewPool("test", work, workerCount, sleepTime, context.Background())

	cancel, err := p.Start()
	defer cancel()
	assert.NoError(t, err)
	assert.NotNil(t, cancel)

	go func() {
		time.Sleep(time.Millisecond * 400)
		cancel()
	}()

	select {
	case <-p.Done():
		t.Fatalf("should have timed out before done channel was closed")
	case <-time.After(time.Millisecond * 200):
	}
}
