# gopool

[![Build Status](https://travis-ci.org/TomWright/gopool.svg?branch=master)](https://travis-ci.org/TomWright/gopool)
[![codecov](https://codecov.io/gh/TomWright/gopool/branch/master/graph/badge.svg)](https://codecov.io/gh/TomWright/gopool)

Go Pool is a package designed to make it simple and easy to set up and get notified about single workers, as well as group those workers into pools and modify the worker count on the fly.

## Installation

```
go get -u github.com/tomwright/gopool
```

## Examples

* [Name Printer using a worker pool](/examples/nameprinterpool)

### Edge case examples

* [Name Printer](/examples/nameprinter)
* [Name Printer (Process Reuse)](/examples/nameprinterreuse)