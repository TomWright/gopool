package gopool

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

func (status ProcessStatus) IsRunning() bool {
	switch status {
	case ProcessStopped:
		return false
	case ProcessStarting:
		return true
	case ProcessRunning:
		return true
	case ProcessStopping:
		return false
	case ProcessFinished:
		return false
	}

	return false
}

const (
	ProcessStopped  ProcessStatus = iota
	ProcessStarting
	ProcessRunning
	ProcessStopping
	ProcessFinished
)
