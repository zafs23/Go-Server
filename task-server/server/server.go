package server

import (
	"fmt"
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
		go HandleConnection(conn)

	}
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Started executing tasks from %v", conn.RemoteAddr())

	

}
