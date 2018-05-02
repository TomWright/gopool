package main

import (
	"github.com/tomwright/gopool"
	"fmt"
	"os"
	"errors"
)

func main() {
	var pNames *chan string

	// we have a list of names and we want to print them to the screen
	names := make(chan string, 10)

	pNames = &names

	for _, v := range []string{"Tom", "Jim", "Jess", "Holly", "Frank"} {
		names <- v
	}

	// create a process to do so
	p := gopool.NewProcess("name-printer-min-3-chars", func(process *gopool.Process, commands <-chan gopool.ProcessCommand) error {
		// inside our func, we want to keep running forever, until either:
		// - the names channel is closed
		// - a stop command is passed in
		// - a name with less than 3 characters is given
		for {
			select {
			case cmd := <-commands: // take a command from the commands channel
				if cmd == gopool.StopProcessCommand {
					return nil
				}
			case name, open := <-*pNames: // take a name from the names channel
				if ! open {
					return nil
				}
				if len(name) < 3 {
					return errors.New(fmt.Sprintf("invalid name: %s. must be at least 3 characters", name))
				}
				fmt.Println(name)
			}
		}
		return nil
	})

	finChan := p.FinishedChan()
	errChan := p.ErrorChan()

	// start the process
	p.Start()

	// add more names to be printed
	for _, v := range []string{"Joe", "Amelia", "Tony", "Elizabeth", "Francesca"} {
		names <- v
	}

	// stop adding names to be printed
	close(names)

	// wait for the process to finish or error
	select {
	case <-finChan:
	case err := <-errChan:
		fmt.Println("Received error: " + err.Error())
	}

	// start the process up again with a different name list
	names = make(chan string)
	pNames = &names
	p.Start()

	names <- "Tom"
	names <- "a"

	// stop adding names to be printed
	close(names)

	// wait for the process to finish or error
	select {
	case <-finChan:
	case err := <-errChan:
		fmt.Println("Received error: " + err.Error())
	}

	os.Exit(0)
}
