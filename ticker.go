package gopool

import "time"

func newTicker(duration time.Duration) *ticker {
	t := new(ticker)
	t.d = duration
	return t
}

type ticker struct {
	T *time.Ticker
	d time.Duration
}

// Start creates a new ticker with the preferred duration
// Note that this does NOT close the previous ticker if one exists
func (t *ticker) Start() {
	t.T = time.NewTicker(t.d)
}

// Stop stops the ticker
func (t *ticker) Stop() {
	t.T.Stop()
}
