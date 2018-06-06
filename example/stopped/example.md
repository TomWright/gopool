# Cancelling a Pool

This is an example how to stop a pool without waiting for it's jobs to be completed.

The full code source can be found at [/example/stopped/pool.go](/example/stopped/pool.go)

The example:
- Creates a pool
- Starts the pool
- Gives some jobs to the workers
- Stops for the pool
- Checks to see if any error occurred

## Running the example

The command:
```
go run example/stopped/pool.go
```

The output:
```
Hello Tom
pool name-printer-pool was stopped due to an error: context canceled
```
