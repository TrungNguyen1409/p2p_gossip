package p2p

import (
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/common"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
	pb "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/proto"
)

type GossipNode struct {
	p2pAddress          string
	peers               map[string]struct{}
	peersMutex          sync.RWMutex
	messageCache        map[string]struct{}
	fanout              int
	gossipInterval      time.Duration
	announceMsgChan     chan enum.AnnounceMsg
	notificationMsgChan chan enum.NotificationMsg
	datatypeMapper      *common.DatatypeMapper
	bootstrapURL        string
}

const (
	fanout         = 2
	gossipInterval = 5 * time.Second
)

func NewGossipNode(p2pAddress string, initialPeers []string, announceMsgChan chan enum.AnnounceMsg, notificationMsgChan chan enum.NotificationMsg, datatypeMapper *common.DatatypeMapper, bootstrapURL string) *GossipNode {
	peers := make(map[string]struct{})
	for _, peer := range initialPeers {
		peers[peer] = struct{}{}
	}

	return &GossipNode{
		p2pAddress:          p2pAddress,
		peers:               peers,
		announceMsgChan:     announceMsgChan,
		notificationMsgChan: notificationMsgChan,
		messageCache:        make(map[string]struct{}),
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
		logger.ErrorF("Failed to start listener for p2p server: %v\n\n", err)
		logger.Info("Default port not available, finding available port...")
		ln, err = net.Listen("tcp", "localhost:0")
		if err != nil {
			logger.FatalF("failed to find an available port: %v", err)
		}
	}

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
		if err1 != nil {
			logger.ErrorF("Failed to accept connection: %v", err)
			continue
		}
		wg.Add(1)
		go node.HandleConnection(conn)
	}
}

// listenAnnounceMessage: listen to announce message and gossip it away. ahh this function is to take announceMsg from the channel (used or intrad node call)
// currently cannot distinguish between announce and notify
func (node *GossipNode) listenAnnounceMessage(announceMsgChan chan enum.AnnounceMsg) {
	logger := logging.NewCustomLogger()

	for {
		msg, ok := <-announceMsgChan
		if !ok {
			logger.Info("Channel closed, exiting loop")
			return
		} else {
			logger.InfoF("P2P Server: Received a message: %+v\n", msg)

			gossipMsg := &pb.GossipMessage{
				MessageId: generate16BitHash(uuid.New().String()),
				Payload:   []byte(msg.Data), // Assuming `Content` is a field in `AnnounceMsg`
				From:      node.p2pAddress,  // Use the node's own address
				Type:      int32(msg.DataType),
				Ttl:       int32(msg.TTL), // Assuming `Type` is a field in `AnnounceMsg`
			}

			node.gossip(gossipMsg)
		}
	}
}

func (node *GossipNode) HandleConnection(conn net.Conn) {
	logger := logging.NewCustomLogger()

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

	node.handleGossipMessage(msg)
}

func (node *GossipNode) handleGossipMessage(msg *pb.GossipMessage) {
	logger := logging.NewCustomLogger()
	//check whether node has demanded Notify by checking whether incoming message has type of Notification 502
	// testing client is currently sending announce message, not notification (fix client a bit)
	// notification is not being send from peer to peer but from gossip to module
	//how to know : peer -----announce----> peer ----- notification ----> module
	//  fix code according to realization above
	logger.InfoF("Received message: %s", string(msg.Payload))

	if _, seen := node.messageCache[msg.MessageId]; seen {
		logger.Info("duplicated message")
		return
	}

	switch msg.Type {
	case int32(enum.GossipAnnounce):
		logger.Info("Receiving AnnounceMessage")
		if node.datatypeMapper.CheckNotify(int(msg.Type)) {
			logger.InfoF("Notification Message found: %s", msg.Type)

			newNotificationMsg := enum.NotificationMsg{
				MessageID: msg.MessageId,
				DataType:  enum.Datatype(enum.GossipNotification),
				Data:      string(msg.Payload),
			}

			node.notificationMsgChan <- newNotificationMsg
			logger.Info("New NotificationMsg added to channel")
		}
		node.messageCache[msg.MessageId] = struct{}{}
		logger.InfoF("New Message saved in Cache with ID: %s", msg.MessageId)
		msg.Ttl -= 1
		node.gossip(msg)

	case int32(enum.PeerAnnounce):
		logger.Debug("Receiving PeerAnnounce : New Peer is Joining")
	case int32(enum.PeerLeave):
		logger.Debug("Receiving PeerLeave : A Peer is leaving")
	case int32(enum.PeerListRequest):
		logger.Debug("Receiving PeerListRequest: A Peer is ")
	case int32(enum.PeerListResponse):
		logger.Debug("Peer returning peer")
	default:
		logger.Debug("default")
	}

}
