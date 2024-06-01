package tests

import (
	"fmt"
	"github.com/robfig/config"
	"github.com/stretchr/testify/require"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/protocols/api"
	"log"
	"net"
	"sync"
	"testing"
)

func Start(t *testing.T) {
	configFile, readErr := config.ReadDefault("../configs/config.ini")
	require.NoError(t, readErr)

	apiAddress, parseErr := configFile.String("gossip", "api_address_test")
	require.NoError(t, parseErr)

	var wg sync.WaitGroup

	go startServer(t, apiAddress, &wg)

	// Wait for all goroutines to finish
	wg.Wait()
}

func startServer(t *testing.T, apiAddress string, wg *sync.WaitGroup) {
	listener, listenerErr := net.Listen("tcp", apiAddress)
	require.NoError(t, listenerErr)

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatalf("Error closing TCP server: %v", err)
		}
	}(listener)

	fmt.Println("Server is listening on", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Increment the WaitGroup counter
		wg.Add(1)

		// Handle this request in a different goroutine
		go func(conn net.Conn) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			api.GossipHandler(conn)
		}(conn)
	}
}
