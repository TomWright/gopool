package gopool

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrPoolAlreadyRunning      = Error("pool already running")
	ErrPoolHasNilSleepTimeFunc = Error("pool has a nil sleep time func")
)
