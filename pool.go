package gopool

import (
	"sync/atomic"
	"fmt"
	"errors"
	"math"
	"time"
)

// NewPool returns a new pool
func NewPool(id string, process func(process *Process, commands <-chan ProcessCommand) error) *Pool {
	p := new(Pool)
	p.id = id
	p.processes = make([]*Process, 0)
	p.process = process
	p.processCounter = 0
	p.status = PoolStopped
	p.SetEnsureProcessCountDuration(time.Second * 30)
	p.SetDesiredProcessCount(func(pool *Pool) uint64 { return 1 })
	return p
}

// Pool represents a group of processes performing the same task simultaneously
type Pool struct {
	id                         string
	processes                  []*Process
	process                    func(process *Process, commands <-chan ProcessCommand) error
	processCounter             uint64
	minProcesses               uint64
	maxProcesses               uint64
	desiredProcessCount        func(pool *Pool) uint64
	status                     PoolStatus
	ensureProcessCountDuration time.Duration
}

// ID returns the pool's id
func (p Pool) ID() string {
	return p.id
}

// Processes returns all of the processes belonging to this pool
func (p Pool) Processes() []*Process {
	return p.processes
}

// ProcessCount returns the current number of processes belonging to this pool
func (p Pool) ProcessCount() int {
	return len(p.processes)
}

func (p *Pool) getNextProcessID() string {
	counter := atomic.AddUint64(&p.processCounter, 1)
	return fmt.Sprintf("%s_%d", p.ID(), counter)
}

// SetDesiredProcessCount defines the way the the pool chooses how many processes should be running
func (p *Pool) SetEnsureProcessCountDuration(duration time.Duration) *Pool {
	p.ensureProcessCountDuration = duration
	return p
}

// SetDesiredProcessCount defines the way the the pool chooses how many processes should be running
func (p *Pool) SetDesiredProcessCount(desiredProcessCount func(pool *Pool) uint64) *Pool {
	p.desiredProcessCount = desiredProcessCount
	return p
}

// AddXProcesses adds X number of processes to the pool
func (p *Pool) AddXProcesses(processesToAdd int) *Pool {
	for x := 0; x < processesToAdd; x++ {
		p.AddProcess()
	}
	return p
}

// RemoveXProcesses removes X number of processes from the pool
func (p *Pool) RemoveXProcesses(processesToRemove int) *Pool {
	for x := 0; x < processesToRemove; x++ {
		if len(p.processes) > 0 {
			p.RemoveProcess(p.processes[0])
		}
	}
	return p
}

// AddProcess spawns a new process under this pool
func (p *Pool) AddProcess() *Process {
	process := NewProcess(p.getNextProcessID(), p.process)
	process.SetPool(p)
	p.processes = append(p.processes, process)
	if p.status.IsRunning() {
		process.Start()
	}
	return process
}

// RemoveProcess stops and then removes the specified process from this pool
func (p *Pool) RemoveProcess(process *Process) *Pool {
	for k, v := range p.processes {
		if v.ID() == process.ID() {
			if ! process.Status().IsRunning() {
				process.Stop()
			}
			p.processes = append(p.processes[:k], p.processes[k+1:]...)
			break
		}
	}

	return p
}

// SetProcessLimits sets the minimum and maximum process counts
func (p *Pool) SetProcessLimits(min uint64, max uint64) *Pool {
	p.minProcesses = min
	p.maxProcesses = max
	return p
}

// Start ensures there are the correct amount of processes and sets the pool status
func (p *Pool) Start() error {
	if p.status != PoolStopped && p.status != PoolFinished {
		return errors.New(fmt.Sprintf("pool is not stopped: %s", p.status.String()))
	}
	p.status = PoolStarting

	go func(p *Pool) {
		for {
			if ! p.status.IsRunning() {
				return
			}
			p.ensureProcessCount()
			time.Sleep(p.ensureProcessCountDuration)
		}
	}(p)

	p.status = PoolRunning

	return nil
}

// Start ensures there are the correct amount of processes and sets the pool status
func (p *Pool) Stop() error {
	p.status = PoolStopping
	p.status = PoolStopped
	p.RemoveXProcesses(p.ProcessCount())
	return nil
}

func (p *Pool) ensureProcessCount() {
	desiredProcessCount := p.desiredProcessCount(p)
	if desiredProcessCount < p.minProcesses && p.minProcesses >= 0 {
		desiredProcessCount = p.minProcesses
	}
	if desiredProcessCount > p.maxProcesses && p.maxProcesses > 0 {
		desiredProcessCount = p.maxProcesses
	}

	diff := float64(p.ProcessCount()) - float64(desiredProcessCount)

	if diff > 0 {
		p.RemoveXProcesses(int(diff))
	} else if diff < 0 {
		p.AddXProcesses(int(math.Abs(diff)))
	}
}
