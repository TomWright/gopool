package gopool

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestPoolStatus_String(t *testing.T) {
	a := assert.New(t)

	a.Equal("stopped", PoolStopped.String())
	a.Equal("starting", PoolStarting.String())
	a.Equal("running", PoolRunning.String())
	a.Equal("stopping", PoolStopping.String())
	a.Equal("finished", PoolFinished.String())

	var unknown PoolStatus
	a.Equal("unknown", unknown.String())
}

func TestPoolStatus_IsRunning(t *testing.T) {
	a := assert.New(t)

	a.False(PoolStopped.IsRunning())
	a.True(PoolStarting.IsRunning())
	a.True(PoolRunning.IsRunning())
	a.False(PoolStopping.IsRunning())
	a.False(PoolFinished.IsRunning())

	var unknown PoolStatus
	a.False(unknown.IsRunning())
}
