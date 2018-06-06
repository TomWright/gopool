# Pool

This is an example of how a basic worker pool may work.

The full code source can be found at [/example/pool/pool.go](/example/pool/pool.go)

The example:
- Creates a pool
- Starts the pool
- Gives some jobs to the workers
- Sleeps for a second to let the jobs run
- Checks to see if any error occurred

## Running the example

The command:
```
go run example/pool/pool.go
```

The output:
```
Hello Frank
Hello Joe
Hello Jess
Hello Tom

```
