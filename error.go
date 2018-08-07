package gopool

type Error string

func (e Error) Error() string {
	return e.String()
}

func (e Error) String() string {
	return string(e)
}

const (
	ErrPoolAlreadyRunning      = Error("pool already running")
	ErrPoolHasNilSleepTimeFunc = Error("pool has a nil sleep time func")
)
