package main

import (
	"os"
	"sync"
	"github.com/tomwright/gopool"
	"fmt"
	"time"
	"math/rand"
)

func main() {
	names := make(chan string, 100)
	wg := sync.WaitGroup{}

	// create a pool with 3 workers
	p := InitPool(3, names, &wg)

	// add some names to the buffered chan
	for _, v := range []string{"Tom", "Jim", "Jess", "Holly", "Frank"} {
		wg.Add(1)
		names <- v
	}

	// start the workers
	p.Start()

	// add more names to be printed
	for _, v := range []string{"Joe", "Amelia", "Tony", "Elizabeth", "Francesca", "Florence", "Jack", "Felix", "Marzia", "Mark"} {
		wg.Add(1)
		names <- v
	}

	// wait for all of the names to be printed
	wg.Wait()

	// we have no more names to print
	// we can either
	close(names) // close the names chan which will close the workers when no more items can be read

	// or we can
	p.Stop() // stop the pool, which passes a stop command to each of the workers and stops them immediately

	os.Exit(0)
}

// InitPool defines a worker pool that reads from a single channel
func InitPool(workers uint64, names <-chan string, wg *sync.WaitGroup) *gopool.Pool {
	p := gopool.NewPool("name-printer", func(process *gopool.Process, commands <-chan gopool.ProcessCommand) error {
		// we want our worker to keep running indefinitely, until:
		// - the names channel is closed
		// - a stop command is passed in
		for {
			select {
			case cmd := <-commands: // take a command from the commands channel
				if cmd == gopool.StopProcessCommand {
					// a stop command was received
					return nil
				}
			case name, open := <-names: // take a name from the names channel
				if ! open {
					// names channel was closed
					// we can't do any work now so return and finish
					return nil
				}
				// print the name with the process id
				fmt.Println(process.ID(), name)

				wg.Done()

				// sleep for a random amount of time to demonstrate workers running simultaneously
				time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			}
		}
	}).
		SetDesiredProcessCount(func(pool *gopool.Pool) uint64 {
		return workers
	})

	return p
}
