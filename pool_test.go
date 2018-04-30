package gopool

import (
	"testing"
	"fmt"
	"time"
	"github.com/stretchr/testify/assert"
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
	p.SetProcessManagerTicker(time.NewTicker(time.Millisecond * 400))
	p.SetDesiredProcessCount(func(pool *Pool) uint64 { return 1 })
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

	p.SetDesiredProcessCount(func(pool *Pool) uint64 { return 5 })
	err = p.Start()
	a.NoError(err)
	time.Sleep(time.Millisecond * 600)
	a.Equal(5, p.ProcessCount())
}
