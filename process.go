package gopool

import (
	"errors"
	"fmt"
)

type ProcessStatus uint32

func (status ProcessStatus) String() string {
	switch status {
	case ProcessStopped:
		return "stopped"
	case ProcessStarting:
		return "starting"
	case ProcessRunning:
		return "running"
	case ProcessStopping:
		return "stopping"
	case ProcessFinished:
		return "finished"
	}

	return "unknown"
}

const (
	ProcessStopped  ProcessStatus = iota
	ProcessStarting
	ProcessRunning
	ProcessStopping
	ProcessFinished
)

func NewProcess(id string, process func(commands <-chan ProcessCommand) error) *Process {
	p := new(Process)
	p.id = id
	p.status = ProcessStopped
	p.process = process
	p.commandChan = make(chan ProcessCommand, 1)
	p.finishedChan = make(chan bool, 1)
	p.errorChan = make(chan error, 1)
	return p
}

type Process struct {
	id           string
	pool         *Pool
	process      func(commands <-chan ProcessCommand) error
	status       ProcessStatus
	commandChan  chan ProcessCommand
	finishedChan chan bool
	errorChan    chan error
}

func (p Process) FinishedChan() chan bool {
	return p.finishedChan
}

func (p Process) ErrorChan() chan error {
	return p.errorChan
}

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

	err := p.process(p.commandChan)
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
