// package test

// import (
// 	"bufio"
// 	"encoding/json"
// 	"fmt"
// 	"go-task-server/task-server/server"
// 	"go-task-server/task-server/tasks"
// 	"net"
// 	"strings"
// 	"sync"
// 	"testing"
// 	"time"
// )

package tasks

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

// // test StartListener
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
	fmt.Println("Listener has been successfully shut down")
}

// Test HandleTask with a mock connection
// func TestHandleTask(t *testing.T) {

// 	// use an in-memory connection using net.Pipe() to simulate mock tcp connection

// 	serverConn, clientConn := net.Pipe()
// 	defer serverConn.Close()
// 	defer clientConn.Close()

// 	// test valid request
// 	//validTask := `{"command":["/bin/echo", "Handling valid task request"],"timeout":1000}` + "\n"

// 	// check if can handle a simple and valid task request
// 	validTaskRequest := tasks.TaskRequest{
// 		Command: []string{"/bin/echo", "Handling valid task request"},
// 		Timeout: 1000,
// 	}

// 	taskRequestBytes, _ := json.Marshal(validTaskRequest)

// 	//done := make(chan struct{})

// 	go func() {
// 		fmt.Printf("Hello from the writer!\n")
// 		_, err := clientConn.Write([]byte(string(taskRequestBytes) + "\n"))
// 		if err != nil {
// 			fmt.Printf("Failed to write to pipe: %v", err)
// 		}
// 		fmt.Println("Done writing!")
// 		clientConn.Close()
// 	}()

// 	go func() {
// 		fmt.Printf("Hello from the reader!\n")
// 		// This should be the function that handles reading from the server connection
// 		tasks.HandleTask(string(taskRequestBytes)+"\n", serverConn)
// 		//close(done)
// 	}()

// 	// start a go routine to simulate server sending task from client side

// 	// call the HandleTask function on the validreuqest and server side connection
// 	//tasks.HandleTask(string(taskRequestBytes)+"\n", serverConn)

// 	//tasks.HandleTask(string(taskRequestBytes)+"\n", serverConn)

// 	// read and check the result from the server side
// 	serverResponse := make([]byte, 1024) // small buffer
// 	res, err := serverConn.Read(serverResponse)

// 	if err != nil {
// 		t.Fatalf("Failed to read from server: %v", err)
// 	}

// 	// // Print raw response for debugging
// 	// fmt.Printf("Raw response from server: %s\n", string(serverResponse[:res]))

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

// TestParseTaskRequest tests if a valid TaskRequest can be parsed correctly.
func TestParseTaskRequest(t *testing.T) {
	taskJSON := `{"command": ["./cmd", "--flag", "arg1"], "timeout": 1000}`
	var task tasks.TaskRequest

	err := json.Unmarshal([]byte(taskJSON), &task)
	if err != nil {
		t.Fatalf("Failed to parse TaskRequest: %v", err)
	}

	if task.Command[0] != "./cmd" || task.Command[1] != "--flag" || task.Command[2] != "arg1" {
		t.Errorf("Parsed command is incorrect: %v", task.Command)
	}

	if task.Timeout != 1000 {
		t.Errorf("Expected timeout to be 1000, got %d", task.Timeout)
	}
}

// TestInvalidTaskRequest tests how HandleTask handles invalid JSON input.
func TestInvalidTaskRequest(t *testing.T) {
	invalidJSON := `{"command": ["./cmd", --flag", "arg1"], "timeout": 1000}`

	conn := &mockConn{}
	tasks.HandleTask(invalidJSON, conn)

	expected := `{"error": "Invalid JSON format"}`
	if !strings.Contains(conn.data, expected) {
		t.Errorf("Expected error response for invalid JSON, got: %s", conn.data)
	}
}

// TestProcessTask tests if a task is processed correctly.
func TestProcessTask(t *testing.T) {
	task := tasks.TaskRequest{
		Command: []string{"echo", "hello"},
		Timeout: 1000,
	}

	result := tasks.ProcessTask(task)

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Output, "hello") {
		t.Errorf("Expected output to contain 'hello', got: %s", result.Output)
	}
}

// TestTaskTimeout tests if a task correctly handles a timeout.
func TestTaskTimeout(t *testing.T) {
	task := tasks.TaskRequest{
		Command: []string{"sleep", "2"},
		Timeout: 500, // 500 ms timeout
	}

	startTime := time.Now()
	result := tasks.ProcessTask(task)
	duration := time.Since(startTime).Milliseconds()

	if result.ExitCode != -1 {
		t.Errorf("Expected exit code -1 for timeout, got %d", result.ExitCode)
	}

	if result.Error != "timeout exceeded" {
		t.Errorf("Expected error 'timeout exceeded', got: %s", result.Error)
	}

	if duration < 500 {
		t.Errorf("Expected task to run for at least 500 ms, ran for %d ms", duration)
	}
}

// Mock connection for testing HandleTask
type mockConn struct {
	data string
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	m.data += string(b)
	return len(b), nil
}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }
