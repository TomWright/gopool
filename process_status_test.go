package gopool

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestProcessStatus_String(t *testing.T) {
	a := assert.New(t)

	a.Equal("stopped", ProcessStopped.String())
	a.Equal("starting", ProcessStarting.String())
	a.Equal("running", ProcessRunning.String())
	a.Equal("stopping", ProcessStopping.String())
	a.Equal("finished", ProcessFinished.String())

	var unknown ProcessStatus = 30
	a.Equal("unknown", unknown.String())
}

func TestProcessStatus_IsRunning(t *testing.T) {
	a := assert.New(t)

	a.False(ProcessStopped.IsRunning())
	a.True(ProcessStarting.IsRunning())
	a.True(ProcessRunning.IsRunning())
	a.False(ProcessStopping.IsRunning())
	a.False(ProcessFinished.IsRunning())

	var unknown ProcessStatus
	a.False(unknown.IsRunning())
}
