package p2p

import (
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
	messageIDCache      []string
	cacheSize           int
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

func NewGossipNode(p2pAddress string, initialPeers []string, announceMsgChan chan enum.AnnounceMsg, notificationMsgChan chan enum.NotificationMsg, datatypeMapper *common.DatatypeMapper, bootstrapURL string, cacheSize int) *GossipNode {
	peers := make(map[string]struct{})
	for _, peer := range initialPeers {
		peers[peer] = struct{}{}
	}

	return &GossipNode{
		p2pAddress:          p2pAddress,
		peers:               peers,
		announceMsgChan:     announceMsgChan,
		notificationMsgChan: notificationMsgChan,
		messageIDCache:      make([]string, 0, cacheSize),
		cacheSize:           cacheSize,
		fanout:              fanout,
		gossipInterval:      gossipInterval,
		datatypeMapper:      datatypeMapper,
		bootstrapURL:        bootstrapURL,
	}
}

func (node *GossipNode) Start() {
	logger := logging.NewCustomLogger()
	if err := node.getInitialPeers(); err != nil {
		logger.FatalF("Failed to register with bootstrapper: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		node.listen(node.p2pAddress, &wg)
	}()

	go func() {
		defer wg.Done()
		node.listenAnnounceMessage(node.announceMsgChan)
	}()

	go func() {
		defer wg.Done()
		node.periodicGossip()
	}()

	go func() {
		defer wg.Done()
		node.periodicBootstrapping()
	}()

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
		logger.ErrorF("Failed to read from connection: %v", err)
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

	logger.InfoF("Received gossip message: %s", string(msg.Payload))

	if node.isMessageCached(strconv.Itoa(int(msg.MessageId))) {
		logger.Info("Duplicated gossip message")
		return
	}

	node.addToCache(strconv.Itoa(int(msg.MessageId)))

	if node.datatypeMapper.CheckNotify(uint16(msg.MessageId), enum.Datatype(msg.Type)) {
		//logger.DebugF("Notification Message found with type: %d", msg.Type)

		newNotificationMsg := enum.NotificationMsg{
			MessageID: uint16(msg.MessageId),
			DataType:  enum.Datatype(msg.Type),
			Data:      string(msg.Payload),
		}

		node.notificationMsgChan <- newNotificationMsg
	}

	msg.Ttl -= 1

	node.gossip(msg, logger)

	/*case int32(enum.PeerAnnounce):
		logger.Debug("Receiving PeerAnnounce : New Peer is Joining")
	case int32(enum.PeerLeave):
		logger.Debug("Receiving PeerLeave : A Peer is leaving")
		// these 2 might not relevant to be handled as gossip message
	case int32(enum.PeerListRequest):
		logger.Debug("Receiving PeerListRequest: A Peer is ")
	case int32(enum.PeerListResponse):
		logger.Debug("Peer returning peer")
	default:
		logger.Debug("default")
	}
		node.notificationMsgChan <- newNotificationMsg
	}

	node.messageCache[msg.MessageId] = struct{}{}
	logger.InfoF("New Message saved in Cache with ID: %s", msg.MessageId)
	msg.Ttl -= 1
	node.gossip(msg, logger)*/
}

func (node *GossipNode) addToCache(msgID string) {
	for _, id := range node.messageIDCache {
		if id == msgID {
			return
		}
	}

	if len(node.messageIDCache) >= node.cacheSize {
		node.messageIDCache = node.messageIDCache[1:]
	}

	node.messageIDCache = append(node.messageIDCache, msgID)

}

func (node *GossipNode) isMessageCached(msgID string) bool {
	for _, id := range node.messageIDCache {
		if id == msgID {
			return true
		}
	}
	return false
}
