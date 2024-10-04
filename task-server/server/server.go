package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/zafs23/Go-Server/task-server/tasks"
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

		wg.Add(1)
		// handle the connection in a concurrently
		go handleConnection(conn, wg)

	}
}

func handleConnection(conn net.Conn, wg *sync.WaitGroup) {

	defer wg.Done()
	defer conn.Close()

	log.Printf("Started executing tasks from %v", conn.RemoteAddr())

	// process the requests
	//scanner := bufio.NewReader(conn)
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		//taskMessage, err := scanner.ReadString('\n')

		taskMessage := scanner.Text()

		if taskMessage == "" {
			log.Printf("Skipping empty lines")
			continue // Skip empty lines, if any
		}

		/*
			no go routine is used here as "After submitting a TaskRequest,
			the scheduler will wait to receive a TaskResult before issuing another new-line terminated TaskRequest."
		*/
		// Process the task
		//log.Printf("Processing task: %s", taskMessage)

		tasks.HandleTask(taskMessage, conn)

	}

	// Check for errors in reading
	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Printf("Failed to read task request from connection %v : %s", conn.RemoteAddr(), err)
	}

	log.Printf("Finished processing tasks from %v", conn.RemoteAddr())

}
