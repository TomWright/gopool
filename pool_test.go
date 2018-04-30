package gopool

import (
	"testing"
	"fmt"
	"time"
	"github.com/stretchr/testify/assert"
	"errors"
)

// Test that the desired process count is acknowledged
func TestPool_SetDesiredProcessCount(t *testing.T) {
	a := assert.New(t)

	// we have a list of names and we want to print them to the screen
	names := make(chan string, 10)

	// create a process to do so
	p := NewPool("name-printer", func(process *Process, commands <-chan ProcessCommand) error {
		// inside our func, we want to keep running forever, until either:
		// - the names channel is closed
		// - a stop command is passed in
		for {
			select {
			case cmd := <-commands: // take a command from the commands channel
				if cmd == StopProcessCommand {
					return nil
				}
			case name, open := <-names: // take a name from the names channel
				if ! open {
					return nil
				}
				fmt.Println(process.ID(), name)
			}
		}
		return nil
	})
	p.SetProcessManagerPollRate(time.Millisecond * 400)
	a.Equal(0, p.ProcessCount())

	// start the pool
	err := p.Start()
	a.NoError(err)
	time.Sleep(time.Millisecond * 600)
	a.Equal(1, p.ProcessCount())

	p.SetDesiredProcessCount(func(pool *Pool) uint64 { return 2 })
	time.Sleep(time.Millisecond * 600)
	a.Equal(2, p.ProcessCount())

	p.SetDesiredProcessCount(func(pool *Pool) uint64 { return 4 })
	time.Sleep(time.Millisecond * 600)
	a.Equal(4, p.ProcessCount())

	p.SetDesiredProcessCount(func(pool *Pool) uint64 { return 3 })
	time.Sleep(time.Millisecond * 600)
	a.Equal(3, p.ProcessCount())

	p.SetDesiredProcessCount(func(pool *Pool) uint64 { return 1 })
	time.Sleep(time.Millisecond * 600)
	a.Equal(1, p.ProcessCount())

	err = p.Stop()
	a.NoError(err)
	time.Sleep(time.Millisecond * 600)
	a.Equal(0, p.ProcessCount())

	// test with process limits
	p.SetDesiredProcessCount(func(pool *Pool) uint64 { return 6 })
	err = p.Start()
	a.NoError(err)
	time.Sleep(time.Millisecond * 600)
	a.Equal(6, p.ProcessCount())

	p.SetDesiredProcessCount(func(pool *Pool) uint64 { return 1 })
	time.Sleep(time.Millisecond * 600)
	a.Equal(1, p.ProcessCount())

	p.SetDesiredProcessCount(func(pool *Pool) uint64 { return 4 })
	time.Sleep(time.Millisecond * 600)
	a.Equal(4, p.ProcessCount())

	p.Stop()
	time.Sleep(time.Millisecond * 600)
	a.Equal(0, p.ProcessCount())
}

func TestPool_Start(t *testing.T) {
	a := assert.New(t)

	// we have a list of names and we want to print them to the screen
	names := make(chan string, 10)

	// create a process to do so
	p := NewPool("name-printer", func(process *Process, commands <-chan ProcessCommand) error {
		// inside our func, we want to keep running forever, until either:
		// - the names channel is closed
		// - a stop command is passed in
		for {
			select {
			case cmd := <-commands: // take a command from the commands channel
				if cmd == StopProcessCommand {
					return nil
				}
			case name, open := <-names: // take a name from the names channel
				if ! open {
					return nil
				}
				fmt.Println(process.ID(), name)
			}
		}
		return nil
	})
	a.Equal(0, p.ProcessCount())

	// start the pool
	err := p.Start()
	a.NoError(err)
	time.Sleep(time.Millisecond * 600)
	a.Equal(1, p.ProcessCount())

	// ensure status is correct
	a.Equal("running", p.Status().String())

	// ensure all processes are running too
	for _, process := range p.Processes() {
		a.True(process.Status().IsRunning())
	}

	err = p.Start()
	a.Error(err, "pool is not stopped: running")

	p.Stop()
}

func TestPool_ErrorChan(t *testing.T) {
	a := assert.New(t)

	names := make(chan string, 10)

	p := NewPool("name-printer", func(process *Process, commands <-chan ProcessCommand) error {
		for {
			select {
			case cmd := <-commands: // take a command from the commands channel
				if cmd == StopProcessCommand {
					return nil
				}
			case name, open := <-names: // take a name from the names channel
				if ! open {
					return nil
				}
				if name == "" {
					return errors.New("name cannot be empty")
				}
				fmt.Println(process.ID(), name)
			}
		}
		return nil
	})
	a.Equal(0, p.ProcessCount())

	// start the pool
	err := p.Start()
	a.NoError(err)
	time.Sleep(time.Millisecond * 600)
	a.Equal(1, p.ProcessCount())

	timeout := time.NewTimer(time.Second * 10)

	names <- "Name 1"
	names <- "Name 2"
	names <- "Name 3"
	names <- ""

	select {
	case <-timeout.C:
		a.Fail("error chan timeout reached. error was expected")
	case err := <-p.ErrorChan():
		a.EqualError(err, "process `name-printer_1` failed: name cannot be empty")
	}

	p.Stop()
}

func TestPool_ErrorChan_MultipleProcesses(t *testing.T) {
	a := assert.New(t)

	names := make(chan string, 10)

	p := NewPool("name-printer", func(process *Process, commands <-chan ProcessCommand) error {
		for {
			select {
			case cmd := <-commands: // take a command from the commands channel
				if cmd == StopProcessCommand {
					return nil
				}
			case name, open := <-names: // take a name from the names channel
				if ! open {
					return nil
				}
				if name == "" {
					return errors.New("name cannot be empty")
				}
				fmt.Println(process.ID(), name)
			}
		}
		return nil
	})
	p.SetDesiredProcessCount(func(pool *Pool) uint64 { return 5 })
	a.Equal(0, p.ProcessCount())

	// start the pool
	err := p.Start()
	a.NoError(err)
	time.Sleep(time.Millisecond * 600)
	a.Equal(5, p.ProcessCount())

	timeout := time.NewTimer(time.Second * 10)

	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- ""
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"
	names <- "Test"

	select {
	case <-timeout.C:
		a.Fail("error chan timeout reached. error was expected")
	case err := <-p.ErrorChan():
		a.Error(err)
	}

	p.Stop()
}

func TestProcess_Pool(t *testing.T) {
	a := assert.New(t)

	names := make(chan string, 10)

	p := NewPool("name-printer", func(process *Process, commands <-chan ProcessCommand) error {
		for {
			select {
			case cmd := <-commands: // take a command from the commands channel
				if cmd == StopProcessCommand {
					return nil
				}
			case name, open := <-names: // take a name from the names channel
				if ! open {
					return nil
				}
				fmt.Println(process.ID(), name)
			}
		}
		return nil
	})
	p.SetDesiredProcessCount(func(pool *Pool) uint64 { return 5 })
	a.Equal(0, p.ProcessCount())

	// start the pool
	err := p.Start()
	a.NoError(err)
	time.Sleep(time.Millisecond * 600)
	a.Equal(5, p.ProcessCount())

	for _, process := range p.Processes() {
		a.Equal(p, process.Pool())
	}

	p.Stop()
}
