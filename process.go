package gopool

import (
	"errors"
	"fmt"
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
}

// ID returns the process id
func (p Process) ID() string {
	return p.id
}

// Pool returns the process pool that this process was spawned by
func (p Process) Pool() *Pool {
	return p.pool
}

// Status returns the process's current status
func (p Process) Status() ProcessStatus {
	return p.status
}

// FinishedChan returns a channel that is written to when the process has finished
func (p Process) FinishedChan() chan bool {
	return p.finishedChan
}

// ErrorChan returns a channel through which errors will be returned
func (p Process) ErrorChan() chan error {
	return p.errorChan
}

// Start will kick off the process
func (p *Process) Start() error {
	if p.status != ProcessStopped && p.status != ProcessFinished {
		return errors.New(fmt.Sprintf("process is not stopped: %s", p.status.String()))
	}

	p.status = ProcessStarting

	go p.run()

	return nil
}

func (p *Process) run() {
	p.status = ProcessRunning

	err := p.process(p, p.commandChan)
	if err != nil {
		p.errorChan <- errors.New(fmt.Sprintf("process `%s` failed: %s", p.id, err))
		p.status = ProcessStopping
	}

	if p.status == ProcessRunning {
		p.status = ProcessFinished
	}
	p.Stop()
}

// Stop will halt the running process by passing a StopProcessingCommand in via the commands channel.
func (p *Process) Stop() error {
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
