package gopool

import (
	"fmt"
	"math"
	"time"
	"sync"
	"errors"
)

// NewPool returns a new pool
func NewPool(id string, process func(process *Process, commands <-chan ProcessCommand) error) *Pool {
	p := new(Pool)
	p.id = id
	p.processes = make([]*Process, 0)
	p.process = process
	p.processCounter = 0
	p.status = PoolStopped
	p.processManagerTicker = newTicker(time.Second * 30)
	p.desiredProcessCount = func(pool *Pool) uint64 { return 1 }
	p.errChan = make(chan error, 100)
	return p
}

// Pool represents a group of processes performing the same task simultaneously
type Pool struct {
	id                         string
	processes                  []*Process
	process                    func(process *Process, commands <-chan ProcessCommand) error
	processCounter             uint64
	desiredProcessCount        func(pool *Pool) uint64
	status                     PoolStatus
	processManagerTicker       *ticker
	ensureProcessCountDuration time.Duration
	errChan                    chan error

	mu sync.Mutex
}

// ID returns the pool's id
func (p Pool) ID() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.id
}

// ErrorChan returns an error channel so as you can receive errors from the processes
func (p Pool) ErrorChan() <-chan error {
	return p.errChan
}

// Start ensures there are the correct amount of processes and sets the pool status
func (p *Pool) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != PoolStopped && p.status != PoolFinished {
		return errors.New(fmt.Sprintf("pool is not stopped: %s", p.status.String()))
	}
	p.status = PoolStarting
	p.startProcessManager()
	p.status = PoolRunning

	return nil
}

// Start ensures there are the correct amount of processes and sets the pool status
func (p *Pool) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.status = PoolStopping
	p.stopProcessManager()
	p.removeXProcesses(len(p.processes))
	p.status = PoolStopped
	return nil
}

// Status returns the processes status
func (p *Pool) Status() PoolStatus {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.status
}

// Processes returns all of the processes belonging to this pool
func (p Pool) Processes() []*Process {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.processes
}

// ProcessCount returns the current number of processes belonging to this pool
func (p Pool) ProcessCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return len(p.processes)
}

// SetProcessManagerPollRate defines how often the process manager should check the process count
func (p *Pool) SetProcessManagerPollRate(duration time.Duration) *Pool {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.processManagerTicker = newTicker(duration)
	return p
}

// SetDesiredProcessCount defines the way the the pool chooses how many processes should be running
func (p *Pool) SetDesiredProcessCount(desiredProcessCount func(pool *Pool) uint64) *Pool {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.desiredProcessCount = desiredProcessCount
	return p
}

func (p *Pool) getNextProcessID() string {
	p.processCounter++
	counter := p.processCounter
	return fmt.Sprintf("%s_%d", p.id, counter)
}

func (p *Pool) startProcessManager() {
	p.processManagerTicker.Start()
	go func() {
		p.mu.Lock()
		p.ensureProcessCount()
		p.mu.Unlock()
	}()
	go func() {
		p.mu.Lock()
		tickerChan := p.processManagerTicker.T.C
		p.mu.Unlock()
		for {
			<-tickerChan
			p.mu.Lock()
			p.ensureProcessCount()
			p.mu.Unlock()
		}
	}()
}

func (p *Pool) stopProcessManager() {
	p.processManagerTicker.Stop()
}

// spawnProcess spawns a new process under this pool
func (p *Pool) spawnProcess() (*Process, error) {
	process := NewProcess(p.getNextProcessID(), p.process)
	process.SetPool(p)
	p.processes = append(p.processes, process)
	errChan := process.ErrorChan()

	if p.status.IsRunning() {
		err := process.Start()
		if err != nil {
			return process, err
		}

		// listen for errors from the process error chan
		// and post them to the pool error chan
		go func(process *Process) {
			for {
				select {
				case err, open := <-errChan:
					if ! open {
						return
					}
					p.errChan <- err
				}
			}
		}(process)
	}
	return process, nil
}

// spawnXProcesses adds X number of processes to the pool
func (p *Pool) spawnXProcesses(processesToAdd int) *Pool {
	for x := 0; x < processesToAdd; x++ {
		p.spawnProcess()
	}
	return p
}

// removeProcess stops and then removes the specified process from this pool
func (p *Pool) removeProcess(process *Process) *Pool {
	for k, v := range p.processes {
		if v.ID() == process.ID() {
			if process.Status().IsRunning() {
				process.Stop()
			}
			p.processes = append(p.processes[:k], p.processes[k+1:]...)
			break
		}
	}

	return p
}

// removeXProcesses removes X number of processes from the pool
func (p *Pool) removeXProcesses(processesToRemove int) *Pool {
	for x := 0; x < processesToRemove; x++ {
		if len(p.processes) > 0 {
			p.removeProcess(p.processes[0])
		}
	}
	return p
}

func (p *Pool) ensureProcessCount() {
	if ! p.status.IsRunning() {
		return
	}

	desiredProcessCount := p.desiredProcessCount(p)
	if desiredProcessCount < 0 {
		desiredProcessCount = 0
	}

	diff := float64(len(p.processes)) - float64(desiredProcessCount)

	if diff > 0 {
		p.removeXProcesses(int(diff))
	} else if diff < 0 {
		p.spawnXProcesses(int(math.Abs(diff)))
	}
}
