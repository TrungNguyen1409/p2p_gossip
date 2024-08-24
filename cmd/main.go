package main

import (
	"strconv"
	"strings"
	"sync"

	"github.com/robfig/config"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/cmd/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/cmd/p2p"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/common"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
)

type Server struct {
	apiServer           *api.Server
	p2pServer           *p2p.GossipNode
	announceMsgChan     chan enum.AnnounceMsg
	notificationMsgChan chan enum.NotifyMsg
	datatypeMapper      *common.DatatypeMapper
}

func NewServer() *Server {
	logger := logging.NewCustomLogger()

	configFile, readErr := config.ReadDefault("configs/config.ini")
	if readErr != nil {
		logger.FatalF("Can not find config.ini %v", readErr)
	}

	difficulty, parseErr := configFile.String("gossip", "difficulty")
	if parseErr != nil {
		logger.FatalF("Can not read from config.ini %v", parseErr)
	}

	difficultyNum, err := strconv.Atoi(difficulty)
	if err != nil {
		logger.FatalF("Error converting difficulty string to int:", err)
	} else {
		enum.Difficulty = strings.Repeat("0", difficultyNum)
	}

	apiAddress, parseErr := configFile.String("gossip", "api_address")
	if parseErr != nil {
		logger.FatalF("Can not read from config.ini %v", parseErr)
	}

	p2pAddress, parseErr := configFile.String("gossip", "p2p_address")
	if parseErr != nil {
		logger.FatalF("Can not read from config.ini %v", parseErr)
	}

	bootstrapperAddress, parseErr := configFile.String("gossip", "bootstrapper_address")
	if parseErr != nil {
		logger.FatalF("Can not read bootstrapper_address from config.ini %v", parseErr)
	}

	announceMsgChan := make(chan enum.AnnounceMsg)
	notificationMsgChan := make(chan enum.NotificationMsg)

	datatypeMapper := common.NewMap()

	apiServer := api.NewServer(apiAddress, announceMsgChan, notificationMsgChan, datatypeMapper)

	p2pServer := p2p.NewGossipNode(p2pAddress, []string{}, announceMsgChan, notificationMsgChan, datatypeMapper, bootstrapperAddress)

	return &Server{apiServer: apiServer, p2pServer: p2pServer, announceMsgChan: announceMsgChan, datatypeMapper: datatypeMapper}

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
