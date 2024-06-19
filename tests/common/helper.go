package tests

import (
	"fmt"
	"github.com/robfig/config"
	"github.com/stretchr/testify/require"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/protocols/api"
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

	ch := make(chan api.AnnounceMsg)

	go startServer(t, apiAddress, &wg, ch)

	// Wait for all goroutines to finish
	wg.Wait()
}

func startServer(t *testing.T, apiAddress string, wg *sync.WaitGroup, announceMsgChan chan api.AnnounceMsg) {
	listener, listenerErr := net.Listen("tcp", apiAddress)
	require.NoError(t, listenerErr)

	logger := logging.NewCustomLogger()

	if listenerErr != nil {
		logger.FatalF("Error starting TCP server: %v", listenerErr)
	}

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			logger.FatalF("Error closing TCP server: %v", err)
		}
	}(listener)

	fmt.Println("API Server is listening on", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.FatalF("Error accepting connection:", err)
			continue
		}

		// Increment the WaitGroup counter
		wg.Add(1)

		// handle this request in a different goroutine
		go func(conn net.Conn) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			logger.Host(conn.LocalAddr().String())
			logger.Client(conn.RemoteAddr().String())
			handler := api.NewHandler(logger)
			handler.Handle(conn, announceMsgChan)
		}(conn)
	}
}
