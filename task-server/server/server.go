package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/zafs23/Go-Server/task-server/tasks"
)

func StartListener(ctx context.Context, wg *sync.WaitGroup, port int) {
	defer wg.Done() // signal the task is done to go routine

	ep := fmt.Sprintf("127.0.0.1:%d", port)

	listener, listenerErr := net.Listen("tcp", ep)
	if listenerErr != nil {
		log.Fatalf("Failed to start listening on port %d: %v", port, listenerErr)
	}

	go func() {
		<-ctx.Done()
		log.Printf("Shutting down listener on port %d", port)
		listener.Close() // force to break the accept() call
	}()

	defer listener.Close()

	log.Printf("Started listening on %s", ep)

	// accept connections in a loop
	for {
		conn, connErr := listener.Accept()
		if connErr != nil {
			select {
			case <-ctx.Done(): // receive cancel() request
				log.Printf("Listener closed on port %d", port)
				return
			default:
				log.Printf("Cannot accept connection on port %d: %v", port, connErr)
				continue
			}
		}
		// conn, err := listener.Accept()
		// if err != nil {
		// 	log.Printf("Cannot accept connection on port %d: %v", port, err)
		// 	continue
		// }

		wg.Add(1)
		// handle the connection concurrently
		go HandleConnection(conn, wg)
		// should not call a wg.wait() here, otherwise it will block each time it accepts new connections

	}
}

// read
func HandleConnection(conn net.Conn, wg *sync.WaitGroup) {

	defer wg.Done()
	defer conn.Close()

	log.Printf("Started executing tasks from %v", conn.RemoteAddr())

	// process the requests
	reader := bufio.NewReaderSize(conn, 8192) // 8KB buffer size
	//reader := bufio.NewReader(conn)
	// we can custom this buffer size

	for {
		//taskMessage, err := scanner.ReadString('\n')
		line, isPrefix, err := reader.ReadLine()

		if err != nil {
			if err == io.EOF {
				log.Printf("Connection closed by client %v", conn.RemoteAddr())
				break
			}
			log.Printf("Failed to read task request from connection %v: %s", conn.RemoteAddr(), err)
			break
		}

		// handle incomplete lines that exceeds the buffer size
		if isPrefix {
			log.Printf("Received an incomplete line (too long) from %v", conn.RemoteAddr())
			continue
			// this handles each request within the buffer size
		}

		taskMessage := string(line)
		if taskMessage == "" {
			log.Printf("Skipping empty lines")
			continue // Skip empty lines, if any
		}

		wg.Add(1)

		// Process the task asynchronously

		go func(taskMsg string, connection net.Conn) {
			defer wg.Done()
			tasks.HandleTask(taskMessage, conn)
		}(taskMessage, conn)

	}
	log.Printf("Finished processing tasks from %v", conn.RemoteAddr())

}
