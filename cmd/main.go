package main

import (
	"fmt"
	"github.com/robfig/config"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/cmd/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/cmd/p2p"
	"log"
	"sync"
)

type Server struct {
	apiServer *api.Server
	p2pServer *p2p.GossipNode
	// channel
}

func NewServer() *Server {
	configFile, readErr := config.ReadDefault("configs/config.ini")
	if readErr != nil {
		log.Fatalf("Can not find config.ini %v", readErr)
	}

	apiAddress, parseErr := configFile.String("gossip", "api_address")
	if parseErr != nil {
		log.Fatalf("Can not read from config.ini %v", parseErr)
	}

	p2pAddress, parseErr := configFile.String("gossip", "p2p_address")
	fmt.Print(p2pAddress)
	if parseErr != nil {
		log.Fatalf("Can not read from config.ini %v", parseErr)
	}

	apiServer := api.NewServer(apiAddress)
	p2pServer := p2p.NewGossipNode(p2pAddress, []string{
		"peer1.example.com:7051",
		"peer2.example.com:7051",
		"peer3.example.com:7051",
	},
	)

	return &Server{apiServer: apiServer, p2pServer: p2pServer}

}

func (s *Server) Start() {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go s.apiServer.Start()
	fmt.Print("starting api server")

	go s.p2pServer.Start()
	fmt.Print("starting p2p server")
	wg.Wait()
}

func main() {
	server := NewServer()
	server.Start()
}
