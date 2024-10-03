package test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go-task-server/task-server/server"
	"go-task-server/task-server/tasks"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// test StartListener
func TestStartListener(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)

	// Start the listenser in a go routine
	go server.StartListener(&wg, 3000)

	time.Sleep(100 * time.Millisecond) // give the server some time to start

	// simulate connection to the server
	conn, err := net.Dial("tcp", "127.0.0.1:3000")

	if err != nil {
		t.Fatalf("Expected to connect to the server, but failed to connect to the server: %v", err)
	}

	defer conn.Close()

	// defer func() {
	// 	fmt.Println("Connection closed:", conn.RemoteAddr())
	// 	// testing if the connection was closed
	// 	conn.Close()
	// }()

	// check if can handle a simple and valid task request
	taskRequest := tasks.TaskRequest{
		Command: []string{"/bin/echo", "Handling valid task request"},
		Timeout: 1000,
	}

	taskRequestBytes, _ := json.Marshal(taskRequest)
	fmt.Fprintf(conn, "%s\n", string(taskRequestBytes))

	// read the output from the server
	serverResponse, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response from the server %v: %v", conn, err)
	}

	// now check if the output matches
	var taskResult tasks.TaskResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(serverResponse)), &taskResult); err != nil {
		t.Fatalf("Error in unmarshaling the response: %v", err)
	}

	if taskResult.Output != "Handling valid task request\n" {
		t.Errorf("expected task result Output 'Handling valid task request', got '%s'", taskResult.Output)
	}

	wg.Done()

}

// Test HandleTask with a mock connection
// func TestHandleTask(t *testing.T) {
// 	// listener, err := net.Listen("tcp", "127.0.0.1:3000")
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }
// 	// defer listener.Close()

// 	// // Start the server in a separate goroutine
// 	// go func() {
// 	// 	conn, err := listener.Accept()
// 	// 	if err != nil {
// 	// 		t.Error(err)
// 	// 		return
// 	// 	}
// 	// 	defer conn.Close()
// 	// }()

// 	// use an in-memory connection using net.Pipe() to simulate mock tcp connection

// 	serverConn, clientConn := net.Pipe()
// 	defer serverConn.Close()
// 	defer clientConn.Close()

// 	// defer func() {
// 	// 	fmt.Println("Connection closed: x", serverConn.RemoteAddr())
// 	// 	serverConn.Close()
// 	// }()

// 	// defer func() {
// 	// 	fmt.Println("Connection closed: y", clientConn.RemoteAddr())
// 	// 	clientConn.Close()
// 	// }()

// 	// test valid request
// 	//validTask := `{"command":["/bin/echo", "Handling valid task request"],"timeout":1000}` + "\n"

// 	// check if can handle a simple and valid task request
// 	validTaskRequest := tasks.TaskRequest{
// 		Command: []string{"/bin/echo", "Handling valid task request"},
// 		Timeout: 1000,
// 	}

// 	taskRequestBytes, _ := json.Marshal(validTaskRequest)
// 	fmt.Printf("here")

// 	// // start a go routine to simulate server sending task from client side
// 	go func() {
// 		defer clientConn.Close()
// 		// write the task to the server side of the pipe
// 		_, err := clientConn.Write([]byte(string(taskRequestBytes) + "\n"))
// 		if err != nil {
// 			fmt.Printf("Failed to write to pipe: %v", err)
// 		}
// 	}()

// 	// call the HandleTask function on the validreuqest and server side connection
// 	tasks.HandleTask(string(taskRequestBytes)+"\n", serverConn)

// 	// read and check the result from the server side
// 	serverResponse := make([]byte, 1024) // small buffer
// 	res, err := serverConn.Read(serverResponse)

// 	if err != nil {
// 		t.Fatalf("Failed to read from server: %v", err)
// 	}

// 	// unmarshal and cross check the response
// 	var taskResult tasks.TaskResult

// 	taskResultErr := json.Unmarshal(serverResponse[:res], &taskResult)

// 	if taskResultErr != nil {
// 		t.Fatalf("Failed to unmarshal response: %v", taskResultErr)
// 	}

// 	// validate the output
// 	if taskResult.Output != "Handling valid task request\n" {
// 		t.Errorf("expected task result Output 'Handling valid task request', got '%s'", taskResult.Output)
// 	}

// }
