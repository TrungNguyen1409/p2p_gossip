package p2p

import (
	"fmt"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/protocols/api"
	"log"
	"sync"

	pb "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/proto"
)

type GossipNode struct {
	p2pAddress      string
	peers           map[string]struct{}
	peersMutex      sync.RWMutex
	messageCache    map[string]struct{}
	announceMsgChan chan api.AnnounceMsg
}

func NewGossipNode(p2pAddress string, initialPeers []string, announceMsgChan chan api.AnnounceMsg) *GossipNode {

	peers := make(map[string]struct{})
	for _, peer := range initialPeers {
		peers[peer] = struct{}{}
	}

	return &GossipNode{
		p2pAddress:      p2pAddress,
		peers:           peers,
		announceMsgChan: announceMsgChan,
		messageCache:    make(map[string]struct{})}
}

func (node *GossipNode) Start() {
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		listen(node.p2pAddress, node.handleMessage, &wg)
	}()

	go func() {
		defer wg.Done()
		listenAnnounceMessage(node.announceMsgChan, &wg)
	}()

	wg.Wait()
}

// where message got handled based on type
func (node *GossipNode) handleMessage(msg *pb.GossipMessage) {
	if _, seen := node.messageCache[string(msg.Payload)]; seen {
		fmt.Println("duplicated message")
		return // already seen this message
	}

	// Process message
	node.messageCache[string(msg.Payload)] = struct{}{}
	log.Printf("Received message: %s", string(msg.Payload))

	// Gossip the message to other peers
	node.peersMutex.RLock()
	defer node.peersMutex.RUnlock()
	for peer := range node.peers {
		if err := send(peer, msg); err != nil {
			log.Printf("Failed to send message to %s: %v", peer, err)
		}
	}
}
