# gopool

[![Build Status](https://travis-ci.org/TomWright/gopool.svg?branch=master)](https://travis-ci.org/TomWright/gopool)
[![codecov](https://codecov.io/gh/TomWright/gopool/branch/master/graph/badge.svg)](https://codecov.io/gh/TomWright/gopool)
[![Documentation](https://godoc.org/github.com/TomWright/gopool?status.svg)](https://godoc.org/github.com/TomWright/gopool)

Go Pool makes it easy to set up and manage pools of background workers.
Not only can they be started and stopped, but they can also be dynamically sized ensuring you have the optimal amount of workers running for your situation.

## Pool Types

### Long running, always alive workers

To start a pool of long running workers who's size will be managed by gopool, you should use the [Pool.Start](https://godoc.org/github.com/TomWright/gopool#Pool.Start) func.

A good use case for this is when you have a message queue that jobs get published to.
Then lets assume you want to always have at least 1 worker consuming jobs from the queue, but you also want to scale the number of consumers according to the current lag.

To do this, just pass in a [WorkerCountFunc](https://godoc.org/github.com/TomWright/gopool#WorkerCountFunc) that returns the number of workers you would want to run at any point in time, and pass in a [SleepTimeFunc](https://godoc.org/github.com/TomWright/gopool#SleepTimeFunc) to customise how often gopool should modify the number of running workers.

E.g. Your [WorkerCountFunc](https://godoc.org/github.com/TomWright/gopool#WorkerCountFunc) could return the `{number of unread items in the queue} / 100`, and the [SleepTimeFunc](https://godoc.org/github.com/TomWright/gopool#SleepTimeFunc) could return `time.Second * 120`.
This would mean that every 2 minutes gopool would resize the pool and ensure that `X / 100` workers are running, where `X` is the number of unread items in the queue.

### Temporary workers, which will die once their designated work has been finished

To start a pool of temporary workers who will die when their work input channel is closed, you should use the [Pool.StartOnce](https://godoc.org/github.com/TomWright/gopool#Pool.StartOnce) func.

To wait for the pool to finish it's work you can use the [Pool.Done](https://godoc.org/github.com/TomWright/gopool#Pool.Done) func, as the channel it returns is closed when the pool's work is done.

## Installation

```
go get -u github.com/tomwright/gopool
```

## Documentation

Documentation can be found at [https://godoc.org/github.com/tomwright/gopool](https://godoc.org/github.com/tomwright/gopool).