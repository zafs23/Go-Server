package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

// TaskRequest represents the incoming task request
type TaskRequest struct {
	Command []string `json:"command"`
	Timeout int      `json:"timeout"`
}

// TaskResult represents the result of the executed task
type TaskResult struct {
	Command    []string `json:"command"`
	ExecutedAt int64    `json:"executed_at"`
	DurationMs float64  `json:"duration_ms"`
	ExitCode   int      `json:"exit_code"`
	Output     string   `json:"output"`
	Error      string   `json:"error"`
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// Read the input from the TCP connection
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Failed to read from connection:", err)
			return
		}

		// Parse the TaskRequest
		var taskReq TaskRequest
		err = json.Unmarshal([]byte(strings.TrimSpace(message)), &taskReq)
		if err != nil {
			log.Println("Failed to unmarshal JSON:", err)
			return
		}

		// Execute the command and respond with TaskResult
		taskResult := executeTask(taskReq)
		resultJSON, err := json.Marshal(taskResult)
		if err != nil {
			log.Println("Failed to marshal result:", err)
			return
		}

		// Send the TaskResult back to the client
		conn.Write(append(resultJSON, '\n'))
	}
}

func executeTask(taskReq TaskRequest) TaskResult {
	start := time.Now()

	// Command execution and context handling for timeout
	ctx := context.Background()
	if taskReq.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(taskReq.Timeout)*time.Millisecond)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, taskReq.Command[0], taskReq.Command[1:]...)

	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	taskResult := TaskResult{
		Command:    taskReq.Command,
		ExecutedAt: start.Unix(),
		DurationMs: float64(duration),
		Output:     string(output),
	}

	// Handle errors
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			taskResult.ExitCode = -1
			taskResult.Error = "timeout exceeded"
		} else {
			if exitErr, ok := err.(*exec.ExitError); ok {
				taskResult.ExitCode = exitErr.ExitCode()
			} else {
				taskResult.ExitCode = -1
			}
			taskResult.Error = err.Error()
		}
	} else {
		taskResult.ExitCode = 0
	}

	return taskResult
}
