package p2p

import (
	"encoding/json"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/common"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/pow"
	pb "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/proto"
)

type GossipNode struct {
	p2pAddress          string
	peers               map[string]struct{}
	peersMutex          sync.RWMutex
	seedNodes           map[string]struct{}
	seedNodesMutex      sync.RWMutex
	isSeedNode          bool
	messageIDCache      []string
	cacheSize           int
	degree              int
	fanout              int
	gossipInterval      time.Duration
	announceMsgChan     chan enum.AnnounceMsg
	notificationMsgChan chan enum.NotificationMsg
	datatypeMapper      *common.DatatypeMapper
	bootstrapURL        string
}

const (
	fanout         = 2
	gossipInterval = enum.GossipInterval
)

func NewGossipNode(
	p2pAddress string,
	initialPeers []string,
	seedNodes []string,
	isSeedNode bool,
	announceMsgChan chan enum.AnnounceMsg,
	notificationMsgChan chan enum.NotificationMsg,
	datatypeMapper *common.DatatypeMapper,
	bootstrapURL string,
	cacheSize int, degree int) *GossipNode {
	peers := make(map[string]struct{})
	seedNodeMap := make(map[string]struct{})

	for _, peer := range initialPeers {
		peers[peer] = struct{}{}
	}
	for _, seedNode := range seedNodes {
		seedNodeMap[seedNode] = struct{}{}
	}
	return &GossipNode{
		p2pAddress:          p2pAddress,
		peers:               peers,
		seedNodes:           seedNodeMap,
		isSeedNode:          isSeedNode,
		announceMsgChan:     announceMsgChan,
		notificationMsgChan: notificationMsgChan,
		messageIDCache:      make([]string, 0, cacheSize),
		cacheSize:           cacheSize,
		degree:              degree,
		fanout:              fanout,
		gossipInterval:      gossipInterval,
		datatypeMapper:      datatypeMapper,
		bootstrapURL:        bootstrapURL,
	}
}

func (node *GossipNode) Start() {
	logger := logging.NewCustomLogger()

	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		node.listen(node.p2pAddress, &wg)
	}()

	time.Sleep(1 * time.Second)

	if err := node.getInitialPeers(); err != nil {
		logger.FatalF("Failed to register with bootstrapper: %v", err)
	}

	node.announceNewPeer()

	node.PrintPeerLists()
	go func() {
		defer wg.Done()
		node.listenAnnounceMessage(node.announceMsgChan)
	}()

	go func() {
		defer wg.Done()
		node.periodicPeerListRequest()
	}()
	/*
		go func() {
			defer wg.Done()
			node.periodicBootstrapping()
		}()
	*/

	go func() {
		defer wg.Done()
		node.sendHeartbeat()
	}()

	wg.Wait()

}

// listen: this function has multiple purposes:
// - listen to messages from other peers
// - gossip the message away
// - check for the type of message that should be propagated, as requested from api of this peer
func (node *GossipNode) listen(p2pAddress string, wg *sync.WaitGroup) {
	logger := logging.NewCustomLogger()

	ln, err := net.Listen("tcp", p2pAddress)
	if err != nil {
		ln, err = net.Listen("tcp", "localhost:0")
		if err != nil {
			logger.FatalF("failed to find an available port: %v", err)
		}
	}
	// line underneath is unnecessary when one node has one corresponding address IP only
	node.p2pAddress = ln.Addr().String()
	logger.InfoF("P2P Server is listening on: %v", ln.Addr())

	if err := node.registerWithBootstrapper(ln.Addr().String()); err != nil {
		logger.FatalF("Failed to register with bootstrapper: %v", err)
	}

	defer func(ln net.Listener) {
		err = ln.Close()
		if err != nil {
			logger.FatalF("Error closing TCP server: %v", err)
		}
	}(ln)

	for {
		conn, err1 := ln.Accept()
		logger.Host(conn.LocalAddr().String())
		logger.Client(conn.RemoteAddr().String())
		if err1 != nil {
			logger.ErrorF("Failed to accept connection: %v", err)
			continue
		}
		wg.Add(1)
		go node.HandleConnection(conn, logger)
	}
}

// listenAnnounceMessage: listen to announce message and gossip it away. this function is to take announceMsg from the channel (used or intrad node call)
func (node *GossipNode) listenAnnounceMessage(announceMsgChan chan enum.AnnounceMsg) {
	logger := logging.NewCustomLogger()

	for {
		msg, ok := <-announceMsgChan
		if !ok {
			logger.Info("Channel closed, exiting loop")
			return
		} else {
			logger.InfoF("P2P Server: Received an Announce message: %+v\n", msg)

			gossipMsg := &pb.GossipMessage{
				MessageId: uint32(generate16BitRandomInteger()),
				Payload:   []byte(msg.Data),
				From:      node.p2pAddress,
				Type:      int32(msg.DataType),
				Ttl:       int32(msg.TTL),
			}

			node.gossip(gossipMsg, logger)
		}
	}
}

func (node *GossipNode) HandleConnection(conn net.Conn, logger *logging.Logger) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.ErrorF("Error closing connection: %v\n", err)
			logger.InfoF("Connection closed!\n----------------------------------\n")
		}
	}(conn)

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			logger.InfoF("Peer %s disconnected", conn.RemoteAddr())
		} else {
			logger.ErrorF("Failed to read from connection: %v", err)
		}
		return
	}
	msg, err := deserialize(buf[:n])
	if err != nil {
		logger.ErrorF("Failed to deserialize message: %v", err)
		return
	}

	if !pow.Validate(msg) {
		logger.Error("Failed to validate nonce")
		return
	}
	node.handleGossipMessage(msg, logger)
}

func (node *GossipNode) handleGossipMessage(msg *pb.GossipMessage, logger *logging.Logger) {

	if node.isMessageCached(strconv.Itoa(int(msg.MessageId))) {
		//logger.InfoF("Duplicated gossip message, ID: %s", strconv.Itoa(int(msg.MessageId)))
		return
	} else {
		node.addToCache(strconv.Itoa(int(msg.MessageId)))
	}

	if node.datatypeMapper.CheckNotify(uint16(msg.MessageId), enum.Datatype(msg.Type)) {
		//logger.DebugF("Notification Message found with type: %d", msg.Type)

		newNotificationMsg := enum.NotificationMsg{
			MessageID: uint16(msg.MessageId),
			DataType:  enum.Datatype(msg.Type),
			Data:      string(msg.Payload),
		}

		node.notificationMsgChan <- newNotificationMsg
	}
	msg.Ttl -= 1 //TODO: check whether there is better place to put this, in gossip() itself for example

	node.handleProtocolMessage(msg, logger)

	node.gossip(msg, logger)

}
func (node *GossipNode) handleProtocolMessage(msg *pb.GossipMessage, logger *logging.Logger) {

	switch msg.Type {

	case int32(enum.PeerJoinAnnounce):
		logger.Debug("Handling PeerJoinAnnounce message")
		peerAddress := string(msg.Payload)
		node.updateByPeerJoin(peerAddress, logger)

	case int32(enum.PeerLeaveAnnounce):
		logger.Debug("Handling PeerLeave message")
		peerAddress := string(msg.Payload)
		node.updateByPeerLeave(peerAddress, logger)

	case int32(enum.PeerListRequest):
		logger.Debug("Handling PeerListRequest message")
		node.respondWithPeerList(msg.From, logger)

	case int32(enum.PeerListResponse):
		logger.Debug("Handling PeerListResponse message")

		var receivedPeers []string
		err := json.Unmarshal(msg.Payload, &receivedPeers)
		if err != nil {
			logger.ErrorF("Failed to unmarshal peer list: %v", err)
			return
		}
		node.updateByPeerListResponse(receivedPeers, logger)

	default:
		logger.DebugF("Unknown P2P message type: %d", msg.Type)
	}
}
