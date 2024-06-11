package api

import (
	"fmt"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/protocols/api"
	"log"
	"net"
	"sync"
)

type Server struct {
	apiAddress      string
	announceMsgChan chan api.AnnounceMsg
}

func NewServer(apiAddress string, announceMsgChan chan api.AnnounceMsg) *Server {
	return &Server{apiAddress: apiAddress, announceMsgChan: announceMsgChan}
}

func (s *Server) Start() {
	var wg sync.WaitGroup

	listen(s.apiAddress, &wg, s.announceMsgChan)

	// Wait for all goroutines to finish
	wg.Wait()
}

func listen(apiAddress string, wg *sync.WaitGroup, announceMsgChan chan api.AnnounceMsg) {
	listener, listenerErr := net.Listen("tcp", apiAddress)
	if listenerErr != nil {
		log.Fatalf("Error starting TCP server: %v", listenerErr)
	}

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatalf("Error closing TCP server: %v", err)
		}
	}(listener)

	fmt.Println("API Server is listening on", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Increment the WaitGroup counter
		wg.Add(1)

		// handle this request in a different goroutine
		go func(conn net.Conn) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			api.Handler(conn, announceMsgChan)
		}(conn)
	}
}
