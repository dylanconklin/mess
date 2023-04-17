package main

import (
	"log"
	"net"
)

func main() {
	// Start and run server
	s := newServer()
	go s.run()
	listener, err := net.Listen("tcp", ":8888")
	// If server can't start
	if err != nil {
		log.Fatalf("unable to start server: %s", err.Error())
	}
	defer listener.Close()
	// Print server status to console with instructions to connect
	log.Println("Server started on Port 8888")
	log.Println("Connect to it by running: mess")
	log.Println("Or: telnet localhost 8888")
	// Accept clients into the server
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %s", err.Error())
			continue
		}
		// Create client goroutine
		client := s.newClient(conn)
		go client.readInput()
	}
}
