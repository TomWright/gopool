package main

import (
	"github.com/tomwright/gopool"
	"context"
	"fmt"
	"time"
)

func main() {
	var nameChan = make(chan string, 10)

	// define the work we want to complete
	var work = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				// if we get a message here, the context has been cancelled
				// and we should return the context err
				return ctx.Err()
			case name, ok := <-nameChan:
				// if we get a message here, either we have a name to do some work with
				// or our job channel has been closed
				if ! ok {
					return nil
				}
				fmt.Printf("Hello %s\n", name)
			}
		}
	}

	// this func returns the number of desired workers to be in use
	var workerCount gopool.WorkerCountFunc = func() uint64 {
		return 3
	}

	// this func returns how long we should sleep in-between making the
	// worker count checks above.s
	var sleepTime gopool.SleepTimeFunc = func() time.Duration {
		return time.Second * 5
	}

	p := gopool.NewPool("name-printer-pool", work, workerCount, sleepTime, context.Background())

	p.Start()

	// the workers are now running in the background...
	// we can pass jobs to them by writing to the nameChan created above
	nameChan <- "Tom"
	nameChan <- "Jess"
	nameChan <- "Frank"
	nameChan <- "Joe"

	// assuming we have no more work for the workers,
	// let's close the jobs chan
	close(nameChan)

	// we can't currently "wait" for all workers to finish working
	// but in this case we can sleep for a second and it's sufficient
	time.Sleep(time.Second)

	// check to see if an error was returned from the worker and log it if found
	err := p.Context().Err()
	if err != nil {
		fmt.Printf("pool %s was stopped due to an error: %s\n", p.ID(), err)
	}

}
