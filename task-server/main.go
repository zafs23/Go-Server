package main

import (
	"log"
	"sync"

	"github.com/zafs23/Go-Server/task-server/server"
)

// declare the ports you want to listen
var ports = []int{3000} // add other ports here

func main() {
	var wg sync.WaitGroup

	for _, port := range ports {
		wg.Add(1)
		//fmt.Printf("%v", port)
		go server.StartListener(&wg, port)
	}

	wg.Wait()
	log.Println("All server tasks are completed")
}
