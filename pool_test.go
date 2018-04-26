package gopool

import (
	"testing"
	"fmt"
	"time"
	"github.com/stretchr/testify/assert"
)

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
	p.Start()
	time.Sleep(time.Millisecond * 20)
	a.Equal(1, p.ProcessCount())

	p.AddProcess()
	time.Sleep(time.Millisecond * 20)
	a.Equal(2, p.ProcessCount())

	p.AddProcess()
	p.AddProcess()
	time.Sleep(time.Millisecond * 20)
	a.Equal(4, p.ProcessCount())

	p.RemoveProcess(p.Processes()[2])
	time.Sleep(time.Millisecond * 20)
	a.Equal(3, p.ProcessCount())

	p.ensureProcessCount()
	time.Sleep(time.Millisecond * 20)
	a.Equal(1, p.ProcessCount())

	p.Stop()
	time.Sleep(time.Millisecond * 20)
	a.Equal(0, p.ProcessCount())
}

func TestPool_EnsureProcessCountDuration(t *testing.T) {
	a := assert.New(t)

	// we have a list of names and we want to print them to the screen
	names := make(chan string, 10)

	var count uint64 = 1
	desiredCount := &count

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
	}).
		SetEnsureProcessCountDuration(time.Millisecond * 100).
		SetDesiredProcessCount(func(pool *Pool) uint64 {
		return *desiredCount
	})

	a.Equal(0, p.ProcessCount())

	// start the pool
	p.Start()
	time.Sleep(time.Millisecond * 20)
	a.Equal(1, p.ProcessCount())

	count = 2
	time.Sleep(time.Millisecond * 500)
	a.Equal(2, p.ProcessCount())

	count = 4
	time.Sleep(time.Millisecond * 500)
	a.Equal(4, p.ProcessCount())

	count = 0
	time.Sleep(time.Millisecond * 500)
	a.Equal(0, p.ProcessCount())

	count = 1
	time.Sleep(time.Millisecond * 500)
	a.Equal(1, p.ProcessCount())

	p.Stop()
	time.Sleep(time.Millisecond * 20)
	a.Equal(0, p.ProcessCount())

	count = 10
	time.Sleep(time.Millisecond * 500)
	a.Equal(0, p.ProcessCount())

	p.Start()
	time.Sleep(time.Millisecond * 20)
	a.Equal(10, p.ProcessCount())

	p.Stop()
	time.Sleep(time.Millisecond * 20)
	a.Equal(0, p.ProcessCount())
}
