package tasks

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"path/filepath"
	"strings"
)

func HandleTask(taskMessage string) {

	// parse the Task Message to create a TaskRequest
	var taskRequest TaskRequest

	taskRequestError := json.Unmarshal([]byte(strings.TrimSpace(taskMessage)), &taskRequest)

	if taskRequestError != nil {
		log.Printf("Error unmarshaling contents of %s: %v", taskMessage, taskRequest)
		return
	}

	// validate the input

	if validationError := validateTaskRequest(taskRequest); validationError != nil {

	}

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

func ExecuteTask(conn net.Conn) {

}
