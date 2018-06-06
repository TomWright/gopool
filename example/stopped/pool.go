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

	cancel := p.Start()

	// the workers are now running in the background...
	// we can pass jobs to them by writing to the nameChan created above
	nameChan <- "Tom"
	nameChan <- "Jess"
	nameChan <- "Frank"
	nameChan <- "Joe"

	// let's sleep for a few microseconds to allow some processing
	time.Sleep(time.Microsecond * 60)
	// then assume some "problem" has occurred
	// and we need to stop the above from processing names immediately
	cancel()

	// wait for the pools context to be done
	<-p.Context().Done()

	// check to see if an error was returned from the worker and log it if found
	err := p.Context().Err()
	if err != nil {
		fmt.Printf("pool %s was stopped due to an error: %s\n", p.ID(), err)
	}

}
