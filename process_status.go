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

const (
	ProcessStopped  ProcessStatus = iota
	ProcessStarting
	ProcessRunning
	ProcessStopping
	ProcessFinished
)
