package test

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/zafs23/Go-Server/task-server/server"
	"github.com/zafs23/Go-Server/task-server/tasks"
)

func TestStartListener(t *testing.T) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go server.StartListener(ctx, &wg, 3001) // Use a different port for testing

	// Give the listener some time to start
	time.Sleep(500 * time.Millisecond)

	// Simulate a client connection
	conn, err := net.Dial("tcp", "127.0.0.1:3001")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}

	// Test sending a message to the server
	_, writeErr := conn.Write([]byte("test task\n"))
	if writeErr != nil {
		t.Fatalf("Failed to write to server: %v", writeErr)
	}

	// Receive a response from the server
	buffer := make([]byte, 1024)
	n, readErr := conn.Read(buffer)
	if readErr != nil {
		t.Fatalf("Failed to read from server: %v", readErr)
	}
	response := string(buffer[:n])
	if response == "" {
		t.Fatalf("Expected a response, got empty string")
	}

	// Close the client connection
	conn.Close()

	// Trigger shutdown
	cancel()

	// Wait for the server to shut down
	wg.Wait()
}

func TestStartListenerHandlesMultipleConnections(t *testing.T) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the server listener in a separate goroutine
	wg.Add(1)
	go server.StartListener(ctx, &wg, 3002) // Use a different port for testing

	// Give the server some time to start
	time.Sleep(500 * time.Millisecond)

	// Number of clients to simulate
	numClients := 5
	var clientWg sync.WaitGroup
	clientWg.Add(numClients)

	// A slice to track if each client received a response
	responses := make([]bool, numClients)

	// Create multiple clients to connect to the server
	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			defer clientWg.Done()

			// Simulate a client connecting to the server
			conn, err := net.Dial("tcp", "127.0.0.1:3002")
			if err != nil {
				t.Errorf("Client %d: Failed to connect to server: %v", clientID, err)
				return
			}
			defer conn.Close()

			// Send a simple task message to the server
			taskMessage := "{\"command\": [\"/bin/echo\", \"Hello from client\"], \"timeout\": 1000}\n"
			_, writeErr := conn.Write([]byte(taskMessage))
			if writeErr != nil {
				t.Errorf("Client %d: Failed to write to server: %v", clientID, writeErr)
				return
			}

			// Read the response
			buffer := make([]byte, 1024)
			conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // Set a read timeout
			n, readErr := conn.Read(buffer)
			if readErr != nil {
				t.Errorf("Client %d: Failed to read from server: %v", clientID, readErr)
				return
			}

			response := string(buffer[:n])
			if response == "" {
				t.Errorf("Client %d: Expected response, but got an empty string", clientID)
			} else {
				responses[clientID] = true
			}
		}(i)
	}

	// Wait for all clients to finish
	clientWg.Wait()

	// Check that all clients received a response
	for i := 0; i < numClients; i++ {
		if !responses[i] {
			t.Errorf("Client %d: Did not receive a response", i)
		}
	}

	// Shutdown the server
	cancel()
	wg.Wait()
}

func TestHandleConnection(t *testing.T) {
	serverS, client := net.Pipe() // Creates an in-memory connection

	var wg sync.WaitGroup
	wg.Add(1)

	// Run handleConnection in a goroutine
	go server.HandleConnection(serverS, &wg)

	// Send a test task
	_, err := client.Write([]byte("{\"command\": [\"/bin/echo\", \"Handling Task\"], \"timeout\": 1000}\n"))
	if err != nil {
		t.Fatalf("Failed to write to connection: %v", err)
	}

	// Read the response
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}
	response := string(buffer[:n])
	if response == "" {
		t.Fatalf("Expected a response, got empty string")
	}

	client.Close()
	wg.Wait()
}

func TestHandleValidTask(t *testing.T) {
	server, client := net.Pipe()

	// Use a simple task request
	taskMessage := "{\"command\": [\"/bin/echo\", \"Handling Task\"], \"timeout\": 1000}"

	go tasks.HandleTask(taskMessage, server)

	// Read the response
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}
	response := string(buffer[:n])
	if !strings.Contains(response, "Handling Task") {
		t.Fatalf("Expected response to contain 'Handling Task', got: %s", response)
	}

	client.Close()
	server.Close()
}

func TestHandleNoTimeOutTask(t *testing.T) {
	server, client := net.Pipe()

	// Use a simple task request without timeout
	taskMessage := "{\"command\": [\"/bin/echo\", \"Handling Task without timeout\"]}"

	go tasks.HandleTask(taskMessage, server)

	// Read the response
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}
	response := string(buffer[:n])
	if !strings.Contains(response, "Handling Task without timeout") {
		t.Fatalf("Expected response to contain 'Handling Task without timeout', got: %s", response)
	}

	client.Close()
	server.Close()
}

func TestHandleTaskValidations(t *testing.T) {
	// Define a set of test cases for various validation errors
	testCases := []struct {
		name          string
		taskMessage   string
		expectedError string
	}{
		{
			name:          "Not Absolute Path",
			taskMessage:   "{\"command\": [\"echo\", \"Handling Task without absolute path\"]}",
			expectedError: "the command is not on the absolute path to the command",
		},
		{
			name:          "Empty Command Array",
			taskMessage:   "{\"command\": [], \"timeout\": 1000}",
			expectedError: "no command is provided",
		},
		{
			name:          "Empty Command",
			taskMessage:   "{\"command\": [\"\"], \"timeout\": 1000}",
			expectedError: "the absolute path to the command is empty",
		},
		{
			name:          "Negative Timeout",
			taskMessage:   "{\"command\": [\"/bin/echo\", \"Hello\"], \"timeout\": -1000}",
			expectedError: "the timeout value should be an positive value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, client := net.Pipe() // Create an in-memory network connection

			go tasks.HandleTask(tc.taskMessage, server)

			// Read the response
			buffer := make([]byte, 1024)
			n, err := client.Read(buffer)
			if err != nil {
				t.Fatalf("Failed to read from connection: %v", err)
			}
			response := string(buffer[:n])

			// Check if the response contains the expected error message
			if !strings.Contains(response, tc.expectedError) {
				t.Fatalf("Expected response to contain '%s', got: %s", tc.expectedError, response)
			}

			client.Close()
			server.Close()
		})
	}
}

func TestHandleInvalidTask(t *testing.T) {
	server, client := net.Pipe()

	// Use a invalid json format task request
	taskMessage := "{\"command\": \"./cmd\", \"--flag\", \"arg1\"], \"timeout\": 1000}"

	go tasks.HandleTask(taskMessage, server)

	// Read the response
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}
	response := string(buffer[:n])
	expectedResponse := `{"error": "Invalid JSON format"}`
	if !strings.Contains(response, expectedResponse) {
		t.Fatalf("Expected response to contain %s, got: %s", expectedResponse, response)
	}

	client.Close()
	server.Close()

}

// TestProcessValidTask tests if a task correctly processes a valid task
func TestProcessValidTask(t *testing.T) {
	taskRequest := tasks.TaskRequest{
		Command: []string{"/bin/echo", "Handling Task"},
		Timeout: 1000,
	}

	result := tasks.ProcessTask(taskRequest)

	if result.ExitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Output, "Handling Task") {
		t.Fatalf("Expected output to contain 'Handling Task', got: %s", result.Output)
	}
}

// TestTaskTimeout tests if a task correctly handles a timeout.
func TestProcessValidTaskTimeout(t *testing.T) {
	taskRequest := tasks.TaskRequest{
		Command: []string{"sleep", "2"},
		Timeout: 500,
	}

	startTime := time.Now()
	result := tasks.ProcessTask(taskRequest)
	duration := time.Since(startTime).Microseconds()

	if result.ExitCode != -1 {
		t.Errorf("Expected exit code -1 for timeout, but got %d", result.ExitCode)
	}

	if result.Error != "timeout exceeded" {
		t.Errorf("Expected error 'timeout exceeded', got: %s", result.Error)
	}

	if duration < 500 {
		t.Errorf("Expected task to run for at least 500 ms, ran for %d ms", duration)
	}

}
