package main

import (
	"github.com/robfig/config"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/cmd/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/cmd/p2p"
	api2 "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/protocols/api"
	"log"
	"sync"
)

type Server struct {
	apiServer       *api.Server
	p2pServer       *p2p.GossipNode
	announceMsgChan chan api2.AnnounceMsg
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
	if parseErr != nil {
		log.Fatalf("Can not read from config.ini %v", parseErr)
	}

	announceMsgChan := make(chan api2.AnnounceMsg)

	apiServer := api.NewServer(apiAddress, announceMsgChan)
	p2pServer := p2p.NewGossipNode(p2pAddress, []string{
		"peer1.example.com:7051",
		"peer2.example.com:7051",
		"peer3.example.com:7051",
	}, announceMsgChan,
	)

	return &Server{apiServer: apiServer, p2pServer: p2pServer, announceMsgChan: announceMsgChan}

}

func (s *Server) Start() {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go s.apiServer.Start()

	go s.p2pServer.Start()
	wg.Wait()
}

func main() {
	server := NewServer()
	server.Start()
}
