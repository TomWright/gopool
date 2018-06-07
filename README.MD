# gopool

[![Build Status](https://travis-ci.org/TomWright/gopool.svg?branch=master)](https://travis-ci.org/TomWright/gopool)
[![codecov](https://codecov.io/gh/TomWright/gopool/branch/master/graph/badge.svg)](https://codecov.io/gh/TomWright/gopool)
[![Documentation](https://godoc.org/github.com/TomWright/gopool?status.svg)](https://godoc.org/github.com/TomWright/gopool)

Go Pool makes it easy to set up and manage pools of background workers.
Not only can they be started and stopped, but they can also be dynamically sized ensuring you have the optimal amount of workers running for your situation.

Think of a message queue that gets batches of messages posted at unknown times.
Rather than just having 1 worker, or a static pool of workers you can dynamically size the worker pool based off of the current lag in the message queue.

This allows you to get through the messages as quickly as possible by spawning up multiple workers, and then once the lag had died down you can kill all but one of the workers.

## Installation

```
go get -u github.com/tomwright/gopool
```

## Examples

* [Pool](/example/pool/example.md)
* [Cancelling a Pool](/example/stopped/example.md)
* [Worker](/example/worker/example.md)
* [Cancelling a Worker](/example/cancelled/example.md)