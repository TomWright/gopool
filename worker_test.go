package gopool_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"context"
	"time"
	"github.com/tomwright/gopool"
	"sync"
	"errors"
)

func TestWorker_ID(t *testing.T) {
	a := assert.New(t)
	w := gopool.NewWorker("id1", nil, context.Background())
	a.Equal("id1", w.ID())
}

func TestWorker_Context(t *testing.T) {
	a := assert.New(t)
	w := gopool.NewWorker("id1", nil, context.Background())
	a.Equal("id1", w.Context().Value("workerId"))
}

func TestWorker_Start(t *testing.T) {
	a := assert.New(t)

	outMu := sync.Mutex{}
	out := make([]int, 0)
	inputChan := make(chan int, 5)
	outputChan := make(chan int, 5)
	var work gopool.WorkFunc = func(ctx context.Context) error {
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

	ctx := context.Background()

	w := gopool.NewWorker("id", work, ctx)

	cancel := w.Start()

	outMu.Lock()
	a.Equal([]int{}, out)
	outMu.Unlock()
	inputChan <- 1
	time.Sleep(time.Microsecond * 500)

	outMu.Lock()
	a.Equal([]int{1}, out)
	outMu.Unlock()
	inputChan <- 1
	time.Sleep(time.Microsecond * 500)

	outMu.Lock()
	a.Equal([]int{1, 1}, out)
	outMu.Unlock()
	inputChan <- 4
	time.Sleep(time.Microsecond * 500)

	outMu.Lock()
	a.Equal([]int{1, 1, 4}, out)
	outMu.Unlock()

	cancel()
	time.Sleep(time.Microsecond * 500)

	inputChan <- 3
	time.Sleep(time.Microsecond * 500)
	outMu.Lock()
	a.Equal([]int{1, 1, 4}, out)
	outMu.Unlock()

	inputChan <- 3
	time.Sleep(time.Microsecond * 50)
	outMu.Lock()
	a.Equal([]int{1, 1, 4}, out)
	outMu.Unlock()
}

var ErrTextExample = errors.New("example error")

func TestWorker_Err(t *testing.T) {
	a := assert.New(t)

	var work gopool.WorkFunc = func(ctx context.Context) error {
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

	w := gopool.NewWorker("id", work, ctx)
	w.Start()

	<-ctx.Done()
	a.EqualError(w.Err(), "example error")
}
