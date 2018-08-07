package gopool_test

import (
	"context"
	"fmt"
	"time"
	"sync"
	"os"
	"github.com/tomwright/gopool"
)

// In this example we start up a pool with 3 long-running workers. We can feed work into these workers at any point
// and even adjust how many workers are running.
// We will be creating a system in which the pool is retrieving numbers from some unknown
// source, does some formatting to them and then adds them to a single log output channel.
func ExamplePool_Start() {
	// This is the channel that the workers will receive the numbers on
	workerInput := make(chan int)

	logOutputChan := make(chan int, 1000)

	var work gopool.WorkFunc = func(ctx context.Context) error {
		fmt.Println("Worker starting")
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Worker finishing")
				return ctx.Err()
			case num, ok := <-workerInput:
				if ! ok {
					// If we get here there are no more numbers being fed in...
					return nil
				}

				num = num * 2
				logOutputChan <- num
			}
		}
	}

	// Define a func to let the pool manager know how many workers to start up.
	// This func can dynamically return different numbers and the pool manager will
	// dynamically resize the pool every X amount of time, where X is specified by the sleepTime func below
	var workerCount gopool.WorkerCountFunc = func() uint64 {
		return 3
	}

	var sleepTime gopool.SleepTimeFunc = func() time.Duration {
		return time.Second
	}

	// Simply create the pool using the vars defined above
	p := gopool.NewPool("number-printer", work, workerCount, sleepTime, context.TODO())

	fmt.Println("Starting pool")
	// Call Start to start the pool in a way that it will automatically restart failed workers
	cancel, err := p.Start()

	// Let's start up a go routine to feed our workers.
	// This will simply add the numbers we provide to the
	// worker input channel
	feederFunc := func(workInput chan<- int, numbers []int) {
		fmt.Println("Starting feeder func")
		for _, i := range numbers {
			workerInput <- i
		}
	}

	// Sleep is here just to assert the example output is correct, and that the 3 workers will already have started
	time.Sleep(time.Millisecond * 500)
	go feederFunc(workerInput, []int{1, 2, 3, 4, 5})

	if err != nil {
		// You should handle this gracefully
		panic(err)
	}
	// Defer the cancel to ensure the pool is stopped
	defer cancel()

	// Let's wait for 500ms and feed in more numbers
	time.Sleep(time.Millisecond * 500)
	go feederFunc(workerInput, []int{6, 7, 8, 9, 10})

	// Since these are long-running workers, the done channel isn't going to be closed when we stop passing in work.
	// Instead we'll sleep for a further 500ms and cancel the pool's context.
	// We can then use the done channel to wait for the pool to be shut down fully.
	// Not that you MUST retrieve the done channel BEFORE you cancel the pool. If you do not do this will will get a nil channel.
	time.Sleep(time.Millisecond * 500)
	fmt.Println("Stopping pool")
	doneChan := p.Done()
	cancel()

	<-doneChan
	// When the done channel is closed is just means that all context cancels have been issued, but it doesn't mean
	// that the workers themselves have acted upon it yet.
	// Let's wait for 500ms to ensure all 3 workers will have exited
	time.Sleep(time.Millisecond * 500)
	fmt.Println("Pool stopped")

	close(logOutputChan)

	// Now let's iterate through the "logged" numbers that had been processed and see the results
	var minOutputNum = -1
	var maxOutputNum = -1
	for i := range logOutputChan {
		if minOutputNum == -1 || i < minOutputNum {
			minOutputNum = i
		}
		if maxOutputNum == -1 || i > maxOutputNum {
			maxOutputNum = i
		}
	}
	fmt.Printf("Min output num: %d\nMax output num: %d\n", minOutputNum, maxOutputNum)

	// Output: Starting pool
	// Worker starting
	// Worker starting
	// Worker starting
	// Starting feeder func
	// Starting feeder func
	// Stopping pool
	// Worker finishing
	// Worker finishing
	// Worker finishing
	// Pool stopped
	// Min output num: 2
	// Max output num: 20
}

// In this example we start up a pool with 3 workers, and a single go routine to feed those workers jobs.
// Once the job feeder has no more work to give it closes the work channel which signals for the workers to quit.
// One the main go routine we simply start all of the above and then wait for it to finish.
// The work in this case is a series of numbers which simply need to be added up.
func ExamplePool_StartOnce() {
	// This is the channel that the workers will receive the numbers on
	workerInput := make(chan int)

	// finalNumber will contain the final number once all given numbers have been summed together
	var finalNumber int
	// finalNumberMu lets us protect the finalNumber from race conditions
	var finalNumberMu sync.Mutex

	var work gopool.WorkFunc = func(ctx context.Context) error {
		// To save us from constantly locking the finalNumber var to stop race conditions, let's
		// add up all the given numbers internally, and then add to the finalNumber when we've
		// got no more work to do
		var workerNumSum int
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case num, ok := <-workerInput:
				if ! ok {
					// If we get here there are no more numbers being fed in...
					// Let's add to the final number!
					finalNumberMu.Lock()
					finalNumber += workerNumSum
					finalNumberMu.Unlock()
					return nil
				}
				// Increment the internal number
				workerNumSum += num
			}
		}
	}

	// Define a func to let the pool manager know how many workers to start up
	var workerCount gopool.WorkerCountFunc = func() uint64 {
		return 3
	}

	// Simply create the pool using the vars defined above
	p := gopool.NewPool("number-summer", work, workerCount, nil, context.TODO())

	fmt.Println("Starting feeder chan")

	// Let's start up a go routine to feed our workers.
	// This will simply add the numbers we provide to the
	// worker input channel, and once all numbers have been
	// added it will close the workInput channel
	go func(workInput chan<- int, numbers []int) {
		for _, i := range numbers {
			workerInput <- i
			fmt.Printf("Number %d was fed\n", i)
		}
		close(workerInput)
		fmt.Println("Work chan was closed")
	}(workerInput, []int{
		1, 2, 3, 4, 5,
		5, 4, 3, 2, 1,
	})

	fmt.Println("Starting pool")
	// Call StartOnce to start the pool in a way that it will finish once all workers have quit
	cancel, err := p.StartOnce()
	if err != nil {
		// You should handle this gracefully
		panic(err)
	}
	// Defer the cancel to ensure the pool is stopped
	defer cancel()

	// Now we wait for either the workers to finish their job, or for a 10 second timeout.
	// If a timeout is hit, we output a message and exit with an error code.
	select {
	case <-p.Done():
	case <-time.After(time.Second * 10):
		fmt.Println("Timeout occurred! Stopping pool")
		cancel()
		os.Exit(9)
	}

	fmt.Printf("Pool finished\nFinal number is %d", finalNumber)

	// Output: Starting feeder chan
	// Starting pool
	// Number 1 was fed
	// Number 2 was fed
	// Number 3 was fed
	// Number 4 was fed
	// Number 5 was fed
	// Number 5 was fed
	// Number 4 was fed
	// Number 3 was fed
	// Number 2 was fed
	// Number 1 was fed
	// Work chan was closed
	// Pool finished
	// Final number is 30
}
