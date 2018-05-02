package gopool

import (
	"errors"
	"fmt"
	"sync"
)

// NewProcess returns a new Process
func NewProcess(id string, process func(process *Process, commands <-chan ProcessCommand) error) *Process {
	p := new(Process)
	p.id = id
	p.status = ProcessStopped
	p.process = process
	p.commandChan = make(chan ProcessCommand, 1)
	p.finishedChan = make(chan bool, 1)
	p.errorChan = make(chan error, 1)
	return p
}

// Process defines a process to run
type Process struct {
	id           string
	pool         *Pool
	process      func(process *Process, commands <-chan ProcessCommand) error
	status       ProcessStatus
	commandChan  chan ProcessCommand
	finishedChan chan bool
	errorChan    chan error

	mu sync.Mutex
}

// ID returns the process id
func (p Process) ID() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.id
}

// Pool returns the process pool that this process was spawned by
func (p Process) Pool() *Pool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.pool
}

// Status returns the process's current status
func (p Process) Status() ProcessStatus {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.status
}

// FinishedChan returns a channel that is written to when the process has finished
func (p Process) FinishedChan() chan bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.finishedChan
}

// ErrorChan returns a channel through which errors will be returned
func (p Process) ErrorChan() chan error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.errorChan
}

// SetPool set's the pool that this process was spawned by
func (p *Process) SetPool(pool *Pool) *Process {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pool = pool
	return p
}

// Start will kick off the process
func (p *Process) Start() error {
	p.mu.Lock()

	if p.status != ProcessStopped && p.status != ProcessFinished {
		p.mu.Unlock()
		return errors.New(fmt.Sprintf("process is not stopped: %s", p.status.String()))
	}

	p.status = ProcessStarting
	p.mu.Unlock()

	p.run()

	return nil
}

func (p *Process) run() {
	p.mu.Lock()
	p.status = ProcessRunning

	errChan := p.errorChan
	cmdChan := p.commandChan

	p.mu.Unlock()

	processFeedbackChan := make(chan interface{})

	go func(p *Process) {
		err := p.process(p, cmdChan)
		processFeedbackChan <- err
	}(p)

	go func() {
		feedback := <-processFeedbackChan
		if err, isErr := feedback.(error); isErr {
			p.mu.Lock()
			p.status = ProcessStopping
			p.mu.Unlock()
			errChan <- errors.New(fmt.Sprintf("process `%s` failed: %s", p.id, err))
		}

		p.mu.Lock()
		if p.status == ProcessRunning {
			p.status = ProcessFinished
		}
		p.stop()
		p.mu.Unlock()
	}()

	return
}

// Stop will halt the running process by passing a StopProcessingCommand in via the commands channel.
func (p *Process) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stop()
}

// Stop will halt the running process by passing a StopProcessingCommand in via the commands channel.
func (p *Process) stop() error {
	if p.status == ProcessStarting || p.status == ProcessRunning {
		p.commandChan <- StopProcessCommand
		p.status = ProcessStopping
		return nil
	}
	if p.status == ProcessFinished || p.status == ProcessStopping {
		p.finishedChan <- true
	}
	if p.status == ProcessStopping {
		p.status = ProcessStopped
	}
	return nil
}
