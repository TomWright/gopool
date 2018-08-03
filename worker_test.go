package gopool_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"context"
	"time"
	. "github.com/tomwright/gopool"
	"sync"
	"errors"
	"fmt"
)

func ExampleWorker() {
	ctx := context.TODO()

	inputChan := make(chan string, 5)
	var work WorkFunc = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case in, ok := <-inputChan:
				if ! ok {
					return nil
				}
				fmt.Print(in)
			}
		}
		return nil
	}

	w := NewWorker("name printer", work, ctx)

	inputChan <- "Tom"
	inputChan <- "Jim"
	inputChan <- "Frank"
	inputChan <- "John"
	inputChan <- "Tony"

	close(inputChan)

	fmt.Print("Start")

	cancel := w.Start()

	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	defer cancel()

	select {
	case <-timer.C:
		panic("worker did not finish in time")
	case <-w.Done():
		fmt.Print("Done")
	}

	// Output: StartTomJimFrankJohnTonyDone
}

func TestWorker_ID(t *testing.T) {
	t.Parallel()

	a := assert.New(t)
	w := NewWorker("id1", nil, context.Background())
	a.Equal("id1", w.ID())
}

func TestWorker_Context(t *testing.T) {
	t.Parallel()

	a := assert.New(t)
	w := NewWorker("id1", nil, context.Background())
	a.Equal("id1", w.Context().Value("workerId"))
}

func TestWorker_Start_Cancel(t *testing.T) {
	t.Parallel()

	outMu := sync.Mutex{}
	out := make([]int, 0)
	inputChan := make(chan int, 5)
	outputChan := make(chan int, 5)
	var work WorkFunc = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case in, ok := <-inputChan:
				if ! ok {
					return nil
				}
				outputChan <- in
			}
		}
		return nil
	}

	outCtx, outCancel := context.WithCancel(context.Background())
	defer outCancel()
	go func(ctx context.Context) {
		for i := range outputChan {
			select {
			case <-ctx.Done():
				return
			default:
				outMu.Lock()
				out = append(out, i)
				outMu.Unlock()
			}
		}
	}(outCtx)

	test := func(expected []int) func() error {
		return func() error {
			outMu.Lock()
			actual := out
			outMu.Unlock()
			if len(actual) != len(expected) {
				return fmt.Errorf("expected (%v), got (%v)", expected, actual)
			}
			for k, v := range actual {
				if expected[k] != v {
					return fmt.Errorf("expected (%v), got (%v)", expected, actual)
				}
			}
			return nil
		}
	}

	ctx := context.Background()

	w := NewWorker("id", work, ctx)

	cancel := w.Start()

	if err := timeoutAfter(time.Second, test([]int{})); err != nil {
		t.Fatal(err)
	}

	inputChan <- 1
	if err := timeoutAfter(time.Second, test([]int{1})); err != nil {
		t.Fatal(err)
	}

	inputChan <- 1
	if err := timeoutAfter(time.Second, test([]int{1, 1})); err != nil {
		t.Fatal(err)
	}

	inputChan <- 4
	if err := timeoutAfter(time.Second, test([]int{1, 1, 4})); err != nil {
		t.Fatal(err)
	}

	inputChan <- 1
	if err := timeoutAfter(time.Second, test([]int{1, 1, 4, 1})); err != nil {
		t.Fatal(err)
	}

	cancel()

	select {
	case <-w.Done():
		break
	case <-time.After(time.Second):
		t.Fatalf("worker did not shut down after %s", time.Second)
	}

	inputChan <- 4
	if err := timeoutAfter(time.Second, test([]int{1, 1, 4, 1, 4})); err == nil {
		t.Fatal(fmt.Errorf("did not expect output to contain 1, 1, 4, 1, 4"))
	}
}

func TestWorker_Start_Finish(t *testing.T) {
	t.Parallel()

	outMu := sync.Mutex{}
	out := make([]int, 0)
	inputChan := make(chan int, 5)
	outputChan := make(chan int, 5)
	var work WorkFunc = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case in, ok := <-inputChan:
				if ! ok {
					return nil
				}
				outputChan <- in
			}
		}
		return nil
	}

	outCtx, outCancel := context.WithCancel(context.Background())
	defer outCancel()
	go func(ctx context.Context) {
		for i := range outputChan {
			select {
			case <-ctx.Done():
				return
			default:
				outMu.Lock()
				out = append(out, i)
				outMu.Unlock()
			}
		}
	}(outCtx)

	test := func(expected []int) func() error {
		return func() error {
			outMu.Lock()
			actual := out
			outMu.Unlock()
			if len(actual) != len(expected) {
				return fmt.Errorf("expected (%v), got (%v)", expected, actual)
			}
			for k, v := range actual {
				if expected[k] != v {
					return fmt.Errorf("expected (%v), got (%v)", expected, actual)
				}
			}
			return nil
		}
	}

	ctx := context.Background()

	w := NewWorker("id", work, ctx)

	w.Start()

	if err := timeoutAfter(time.Second, test([]int{})); err != nil {
		t.Fatal(err)
	}

	inputChan <- 1
	if err := timeoutAfter(time.Second, test([]int{1})); err != nil {
		t.Fatal(err)
	}

	inputChan <- 1
	if err := timeoutAfter(time.Second, test([]int{1, 1})); err != nil {
		t.Fatal(err)
	}

	inputChan <- 4
	if err := timeoutAfter(time.Second, test([]int{1, 1, 4})); err != nil {
		t.Fatal(err)
	}

	inputChan <- 1
	if err := timeoutAfter(time.Second, test([]int{1, 1, 4, 1})); err != nil {
		t.Fatal(err)
	}

	close(inputChan)

	select {
	case <-w.Done():
		break
	case <-time.After(time.Second):
		t.Fatalf("worker did not shut down after closing input chan. waited for %s", time.Second)
	}
}

var ErrTextExample = errors.New("example error")

func TestWorker_Err(t *testing.T) {
	t.Parallel()

	a := assert.New(t)

	var work WorkFunc = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return ErrTextExample
			}
		}
		return nil
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Second)

	w := NewWorker("id", work, ctx)
	w.Start()

	<-ctx.Done()
	a.EqualError(w.Err(), "example error")
}
