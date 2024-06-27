package api

import (
	"net"
	"sync"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/common"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/protocols/api"
)

type Server struct {
	apiAddress      string
	announceMsgChan chan enum.AnnounceMsg
	datatypeMapper  *common.DatatypeMapper
}

func NewServer(apiAddress string, announceMsgChan chan enum.AnnounceMsg, datatypeMapper *common.DatatypeMapper) *Server {
	return &Server{apiAddress: apiAddress, announceMsgChan: announceMsgChan, datatypeMapper: datatypeMapper}
}

func (s *Server) Start() {
	var wg sync.WaitGroup

	listen(s.apiAddress, &wg, s.announceMsgChan, s.datatypeMapper)

	// Wait for all goroutines to finish
	wg.Wait()
}

func listen(apiAddress string, wg *sync.WaitGroup, announceMsgChan chan enum.AnnounceMsg, datatypeMapper *common.DatatypeMapper) {
	listener, listenerErr := net.Listen("tcp", apiAddress)

	logger := logging.NewCustomLogger()

	if listenerErr != nil {
		logger.ErrorF("Error starting TCP server: %v", listenerErr)
		logger.InfoF("Error starting TCP P2P server: %v\n", listenerErr)
		logger.Info("Default port not available, finding available port...")
		listener, listenerErr = net.Listen("tcp", "localhost:0")
		if listenerErr != nil {
			logger.ErrorF("failed to find an available port: %v", listenerErr)
		}
	}

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			logger.FatalF("Error closing TCP server: %v", err)
		}
	}(listener)

	logger.InfoF("API Server is listening on: %v", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.InfoF("Error accepting connection: ", err)
			continue
		}

		// Increment the WaitGroup counter
		wg.Add(1)

		// handle this request in a different goroutine
		go func(conn net.Conn) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			logger.Host(conn.LocalAddr().String())
			logger.Client(conn.RemoteAddr().String())
			handler := api.NewHandler(conn, logger, announceMsgChan, datatypeMapper)
			handler.Handle()
		}(conn)
	}
}
