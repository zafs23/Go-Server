package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func HandleTask(taskMessage string, conn net.Conn) {

	// parse the Task Message to create a TaskRequest
	var taskRequest TaskRequest

	taskRequestError := json.Unmarshal([]byte(strings.TrimSpace(taskMessage)), &taskRequest)

	if taskRequestError != nil {
		log.Printf("Error unmarshaling contents of %s: %v", taskMessage, taskRequest)
		fmt.Fprintf(conn, `{"error": "Invalid JSON format"}\n`)
		return
	}

	// validate the input

	if validationError := validateTaskRequest(taskRequest); validationError != nil {
		log.Printf("Validation failed: %v", validationError)
		fmt.Fprintf(conn, `{"error": "%s"}\n`, validationError.Error())
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func(task TaskRequest) {
		defer wg.Done()
		taskResult := ProcessTask(task)

		processBytes, processError := json.Marshal(taskResult)

		if processError != nil {
			log.Fatalf("Error marshalling result: %v", processError)
			fmt.Fprintf(conn, `{"error": "Failed to process task"}\n`)
		} else {
			fmt.Fprintf(conn, "%s\n", string(processBytes))
		}
	}(taskRequest)

	wg.Wait()

}

func ProcessTask(task TaskRequest) TaskResult {
	startTime := time.Now()

	executedAt := startTime.Unix()

	// creating a default timeout when timeout is not given
	if task.Timeout == nil {
		task.Timeout = create(6000)
	}
	timeOutDuration := time.Duration(*task.Timeout) * time.Millisecond

	taskResult := TaskResult{
		Command:    task.Command,
		ExecutedAt: executedAt,
	}

	cmd := exec.Command(task.Command[0], task.Command[1:]...)
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output

	// run the command in a go routine and send the error message through a channel
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	/*
		if timeout exceeded, kill the process (asumed) and put the status code -1, and error to timeout exceeded
	*/
	case <-time.After(timeOutDuration):
		cmd.Process.Kill()
		taskResult.ExitCode = -1
		taskResult.Error = "timeout exceeded"
	/*
		if the process was failed to run, check the error message
	*/
	case cmdErr := <-done:
		duration := time.Since(startTime).Microseconds()
		taskResult.DurationMs = float64(duration)

		if cmdErr != nil {
			taskResult.ExitCode = -1
			taskResult.Error = cmdErr.Error()
		} else { // there was no error on running the command, capture everything on the output
			taskResult.ExitCode = 0
			taskResult.Output = output.String()
		}

	}

	return taskResult

}

func validateTaskRequest(taskRequest TaskRequest) error {
	// validate if there is only new line
	if len(taskRequest.Command) == 0 {
		return errors.New("no command is provided")
	}
	// check if there is no command in the command array

	if taskRequest.Command[0] == "" {
		return errors.New("the absolute path to the command is empty")
	}

	// check if the command is in the absolute path of the command

	if !filepath.IsAbs(taskRequest.Command[0]) {
		return errors.New("the command is not on the absolute path to the command")
	}

	// check if the time out is negative value

	if taskRequest.Timeout != nil && *taskRequest.Timeout < 0 {
		return errors.New("the timeout value should be an positive value")
	}

	return nil
}

func create(x int) *int {
	return &x
}
