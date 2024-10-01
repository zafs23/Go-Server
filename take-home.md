# Description

Your task is to write a small server that responds to network requests from
another service. The server should listen to requests from localhost:3000 on 
a raw TCP socket and execute commands received from a remote client.

# Requirements

Accept incoming tasks from the scheduler on `127.0.0.1:3000` TCP. Each message
will have the following task request structure:

```json
{
  "command": ["./cmd", "--flag", "argument1", "argument2"],
  "timeout": 500
}
```

We'll refer to this as a TaskRequest.

Command is the ARGV array of arguments, the first of which is the absolute
path to the command being executed.

Timeout is in milliseconds. A timeout of 0 or a missing timeout field means there
is no timeout.

Task requests will be terminated by a new-line. After submitting a TaskRequest, 
the scheduler will wait to receive a TaskResult before issuing another new-line terminated TaskRequest.

Your agent should respond with a TaskResult:

```json
{
  "command": ["./cmd", "--flag", "argument1", "argument2"],
  "executed_at": 0,
  "duration_ms": 0.0,
  "exit_code": 0,
  "output": "",
  "error": "",
}
```

Upon receiving a TaskRequest:

  - Record the time the task was executed and put it in the `executed_at` field.
  - Record the duration of execution and put it in the `duration_ms` field.
  - Record the exit status of the subprocess and put it in the `exit_code` field.
    - If the process failed to execute at all and there is no status code, assign a
      value of `-1` and put the error in the `error` field.
    - If the specified timeout was exceeded, assign a value of `-1` to the and
      assign the string `timeout exceeded` to the `error` field.
  - Capture everything written to STDOUT and put it in the `output` field.

The solution needs to be able to process requests in parallel, meaning that it can accept and start processing 
multiple requests in parallel, without yet having returned the previous results.

The solution must contain unit tests validating the results.

### Environment
* Go must be used to implement the solution.
* Please provide the solution by sharing a GitHub repository (or any other publicly accessible Git repository).