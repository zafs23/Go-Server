package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/zafs23/Go-Server/task-server/server"
)

// declare the ports you want to listen
var ports = []int{3000} // add other ports here

func main() {
	var wg sync.WaitGroup

	// Create a context to handle server graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	for _, port := range ports {
		wg.Add(1)
		//fmt.Printf("%v", port)
		go server.StartListener(ctx, &wg, port)
	}

	// Listen for interrupt signals to gracefully shut down the server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan // Block until an interrupt signal is received
	log.Println("Shutdown signal received")

	// Trigger the server shutdown
	cancel()
	wg.Wait()
	log.Println("All server tasks are completed!")
}
