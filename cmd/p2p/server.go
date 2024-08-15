package p2p

import (
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/common"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
	pb "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/proto"
)

type GossipNode struct {
	p2pAddress      string
	peers           map[string]struct{}
	peersMutex      sync.RWMutex
	messageCache    map[string]struct{}
	fanout          int
	gossipInterval  time.Duration
	announceMsgChan chan enum.AnnounceMsg
	datatypeMapper  *common.DatatypeMapper
	bootstrapURL    string
}

const (
	fanout         = 2
	gossipInterval = 5 * time.Second
)

func NewGossipNode(p2pAddress string, initialPeers []string, announceMsgChan chan enum.AnnounceMsg, datatypeMapper *common.DatatypeMapper, bootstrapURL string) *GossipNode {
	peers := make(map[string]struct{})
	for _, peer := range initialPeers {
		peers[peer] = struct{}{}
	}

	return &GossipNode{
		p2pAddress:      p2pAddress,
		peers:           peers,
		announceMsgChan: announceMsgChan,
		messageCache:    make(map[string]struct{}),
		fanout:          fanout,
		gossipInterval:  gossipInterval,
		datatypeMapper:  datatypeMapper,
		bootstrapURL:    bootstrapURL,
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
		node.listen(node.p2pAddress, node.handleMessage, &wg)
	}()

	go func() {
		defer wg.Done()
		node.listenAnnounceMessage(node.announceMsgChan, &wg)
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
func (node *GossipNode) listen(p2pAddress string, msgHandler func(*pb.GossipMessage), wg *sync.WaitGroup) {
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

	// Register with Bootstrapper
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
		go HandleConnection(conn, msgHandler)
	}
}

// listenAnnounceMessage: listen to announce message and gossip it away.
func (node *GossipNode) listenAnnounceMessage(announceMsgChan chan enum.AnnounceMsg, wg *sync.WaitGroup) {
	logger := logging.NewCustomLogger()

	for {
		msg, ok := <-announceMsgChan
		if !ok {
			// Channel is closed, exit the loop
			logger.Info("Channel closed, exiting loop")
			return
		} else {
			logger.InfoF("P2P Server: Received a message: %+v\n", msg)

			// handle gossip algorithm here!
			gossipMsg := &pb.GossipMessage{
				Payload: []byte("hello"),
				From:    "localhost", // Replace with actual address if needed
			}

			// Add message to cache to prevent reprocessing
			//messageCache[msg.Message] = struct{}{}

			node.gossip(gossipMsg)
		}
	}
}

func HandleConnection(conn net.Conn, msgHandler func(*pb.GossipMessage)) {
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

	msgHandler(msg)
}

func (node *GossipNode) handleMessage(msg *pb.GossipMessage) {
	logger := logging.NewCustomLogger()

	if _, seen := node.messageCache[string(msg.Payload)]; seen {
		logger.Info("duplicated message")
		return // already seen this message
	}

	// Process message
	node.messageCache[string(msg.Payload)] = struct{}{}
	logger.InfoF("Received message: %s", string(msg.Payload))

	// Gossip the message to other peers
	node.peersMutex.RLock()
	defer node.peersMutex.RUnlock()
	for peer := range node.peers {
		if err := send(peer, msg); err != nil {
			logger.ErrorF("Failed to send message to %s: %v", peer, err)
		}
	}
}

func serialize(msg *pb.GossipMessage) ([]byte, error) {
	return proto.Marshal(msg)
}

func deserialize(data []byte) (*pb.GossipMessage, error) {
	var msg pb.GossipMessage
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// send sends a message to a peer
func send(address string, msg *pb.GossipMessage) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	data, err := serialize(msg)
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}
