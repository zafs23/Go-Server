package server

import (
	"bufio"
	"fmt"
	"go-task-server/task-server/tasks"
	"io"
	"log"
	"net"
	"sync"
)

func StartListener(wg *sync.WaitGroup, port int) {
	defer wg.Done() // signal the task is done to go routine

	ep := fmt.Sprintf("127.0.0.1:%d", port)

	listener, err := net.Listen("tcp", ep)
	if err != nil {
		log.Fatalf("Failed to start listening on port %d: %v", port, err)
	}

	defer listener.Close()

	log.Printf("Started listening on %s", ep)

	// accept connections in a loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Cannot accept connection on port %d: %v", port, err)
			continue
		}

		// handle the connection in a concurrently
		go handleConnection(conn)

	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("Started executing tasks from %v", conn.RemoteAddr())

	// process the requests
	//scanner := bufio.NewReader(conn)
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		//taskMessage, err := scanner.ReadString('\n')

		taskMessage := scanner.Text()

		// if err != nil {
		// 	if err == io.EOF {
		// 		log.Fatalf("Connection closed by client %v", conn.RemoteAddr())
		// 		//return // or close the connection gracefully
		// 	}
		// 	log.Fatalf("Failed to read task request from connection %v : %s", conn.RemoteAddr(), err)
		// 	//return

		/*
			no go routine is used here as "After submitting a TaskRequest,
			the scheduler will wait to receive a TaskResult before issuing another new-line terminated TaskRequest."
		*/
		if taskMessage == "" {
			continue // Skip empty lines, if any
		}

		tasks.HandleTask(taskMessage, conn)

	}

	// Check for errors in reading
	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Printf("Failed to read task request from connection %v : %s", conn.RemoteAddr(), err)
	}

	log.Printf("Finished processing tasks from %v", conn.RemoteAddr())

}
