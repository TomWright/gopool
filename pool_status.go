package gopool

type PoolStatus uint32

func (status PoolStatus) String() string {
	switch status {
	case PoolStopped:
		return "stopped"
	case PoolStarting:
		return "starting"
	case PoolRunning:
		return "running"
	case PoolStopping:
		return "stopping"
	case PoolFinished:
		return "finished"
	}

	return "unknown"
}

func (status PoolStatus) IsRunning() bool {
	switch status {
	case PoolStopped:
		return false
	case PoolStarting:
		return true
	case PoolRunning:
		return true
	case PoolStopping:
		return false
	case PoolFinished:
		return false
	}

	return false
}

const (
	PoolStopped  PoolStatus = iota
	PoolStarting
	PoolRunning
	PoolStopping
	PoolFinished
)
