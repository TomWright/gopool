package gopool_test

import (
	"time"
	"fmt"
)

func timeoutAfter(timeout time.Duration, fn func() error) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	var err error
	for {
		// Run the function and save the last error, if any. Exit if no error.
		if err = fn(); err == nil {
			return nil
		}

		// Fail test if timeout occurs.
		// Otherwise wait for initial channel or interval channel.
		select {
		case <-timer.C:
			return fmt.Errorf("%s (%s timeout)", err, timeout)
		case <-ticker.C:
			continue
		}
	}
}
