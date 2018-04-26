package main

import (
	"github.com/tomwright/gopool"
	"fmt"
	"os"
)

func main() {
	// we have a list of names and we want to print them to the screen
	names := make(chan string, 10)
	for _, v := range []string{"Tom", "Jim", "Jess", "Holly", "Frank"} {
		names <- v
	}

	// create a process to do so
	p := gopool.NewProcess("name-printer", func(process *gopool.Process, commands <-chan gopool.ProcessCommand) error {
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
				fmt.Println(name)
			}
		}
		return nil
	})

	// start the process
	p.Start()

	// add more names to be printed
	for _, v := range []string{"Joe", "Amelia", "Tony", "Elizabeth", "Francesca"} {
		names <- v
	}

	// stop adding names to be printed
	close(names)

	// wait for the process to finish
	<-p.FinishedChan()

	os.Exit(0)
}
