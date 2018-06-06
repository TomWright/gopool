# Cancelling a Worker

This is an example how to cancel a worker without waiting for it's jobs to be completed.

The full code source can be found at [/example/cancelled/worker.go](/example/cancelled/worker.go)

The example:
- Creates a worker
- Starts the worker
- Gives some jobs to the worker
- Cancels the worker
- Waits for the worker to finish
- Checks to see if any error occurred

## Running the example

The command:
```
go run example/cancelled/worker.go
```

The output:
```
Hello Tom
Hello Jess
worker worker-2 was stopped due to an error: context canceled
```
