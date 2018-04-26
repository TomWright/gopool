package main

import (
	"github.com/tomwright/gopool"
	"fmt"
	"os"
	"time"
	"math/rand"
)

func main() {
	// we have a list of names and we want to print them to the screen
	names := make(chan string, 10)
	for _, v := range []string{"Tom", "Jim", "Jess", "Holly", "Frank"} {
		names <- v
	}

	// create a process to do so
	p := gopool.NewPool("name-printer", func(process *gopool.Process, commands <-chan gopool.ProcessCommand) error {
		// inside our func, we want to keep running forever, until either:
		// - the names channel is closed
		// - a stop command is passed in
		for {
			select {
			case cmd := <-commands: // take a command from the commands channel
				if cmd == gopool.StopProcessCommand {
					return nil
				}
			case name, open := <-names: // take a name from the names channel
				if ! open {
					return nil
				}
				fmt.Println(process.ID(), name)
				time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			}
		}
		return nil
	}).
		SetDesiredProcessCount(func(pool *gopool.Pool) uint64 {
		return 3 // we want 3 of the above processes running
	})

	// start the process
	p.Start()

	// add more names to be printed
	for _, v := range []string{"Joe", "Amelia", "Tony", "Elizabeth", "Francesca", "Florence", "Jack", "Felix", "Marzia", "Mark"} {
		names <- v
	}

	// stop adding names to be printed
	close(names)

	// wait for the process to finish
	time.Sleep(time.Millisecond * 1000)

	os.Exit(0)
}
