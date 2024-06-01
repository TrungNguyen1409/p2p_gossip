package p2p

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type Server struct {
	p2pAddress string
}

func NewServer(p2pAddress string) *Server {
	return &Server{p2pAddress: p2pAddress}
}

func (s *Server) Start() {
	var wg sync.WaitGroup

	listen(s.p2pAddress, &wg)

	// Wait for all goroutines to finish
	wg.Wait()
}

func listen(p2pAddress string, wg *sync.WaitGroup) {
	listener, listenerErr := net.Listen("tcp", p2pAddress)
	if listenerErr != nil {
		log.Fatalf("Error starting TCP server: %v", listenerErr)
	}

	fmt.Println("P2P Server is listening on", listener.Addr())

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatalf("Error closing TCP server: %v", err)
		}
	}(listener)
}
