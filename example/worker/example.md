# Example Worker

This is an example how to use a basic worker.

The full code source can be found at [/example/worker/worker.go](/example/worker/worker.go)

The example:
- Creates a worker
- Starts the worker
- Gives some jobs to the worker
- Waits for the worker to finish processing the jobs
- Checks to see if any error occurred

## Running the example

The command:
```
go run example/worker/worker.go
```

The output:
```
Hello Tom
Hello Jess
Hello Frank
Hello Joe
```
