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
		taskResult := processTask(task)

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

func processTask(task TaskRequest) TaskResult {
	startTime := time.Now()

	executedAt := startTime.Unix()

	// creating a default timeout when timeout is not given
	const defaultTimeout = 10000
	if task.Timeout == 0 {
		task.Timeout = defaultTimeout
	}
	timeOutDuration := time.Duration(task.Timeout) * time.Millisecond

	taskResult := TaskResult{
		Command:    task.Command,
		ExecutedAt: executedAt,
	}

	// change the taskRequest absolute path to base command from command[0]
	//taskRequest.Command = append([]string{filepath.Base(task.Command[0])}, task.Command[1:]...)

	baseCmd := filepath.Base(task.Command[0])
	cmd := exec.Command(baseCmd, task.Command[1:]...)
	var output, outputError strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &outputError

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
		taskResult.DurationMs = float64(task.Timeout)
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
			taskResult.Error = outputError.String() // when the command ran, but has both error and stdout
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

	if taskRequest.Timeout < 0 {
		return errors.New("the timeout value should be an positive value")
	}

	return nil
}
