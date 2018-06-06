# gopool

[![Build Status](https://travis-ci.org/TomWright/gopool.svg?branch=master)](https://travis-ci.org/TomWright/gopool)
[![codecov](https://codecov.io/gh/TomWright/gopool/branch/master/graph/badge.svg)](https://codecov.io/gh/TomWright/gopool)
[![Documentation](https://godoc.org/github.com/TomWright/gopool?status.svg)](https://godoc.org/github.com/TomWright/gopool)

Go Pool is a package designed to make it simple and easy to set up and get notified about single workers, as well as group those workers into pools and modify the worker count on the fly.

## Installation

```
go get -u github.com/tomwright/gopool
```

## Examples

* [Worker](/example/worker/example.md)
* [Cancelling a Worker](/example/cancelled/example.md)
* [Pool](/example/pool/example.md)
* [Cancelling a Pool](/example/stopped/example.md)