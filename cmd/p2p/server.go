package p2p

import (
	"log"
	"sync"
)

type GossipNode struct {
	p2pAddress   string
	peers        map[string]struct{}
	peersMutex   sync.RWMutex
	messageCache map[string]struct{}
}

func NewGossipNode(p2pAddress string, initialPeers []string) *GossipNode {

	peers := make(map[string]struct{})
	for _, peer := range initialPeers {
		peers[peer] = struct{}{}
	}

	return &GossipNode{
		p2pAddress:   p2pAddress,
		peers:        peers,
		messageCache: make(map[string]struct{})}
}

func (node *GossipNode) Start() {
	var wg sync.WaitGroup

	ListenForMessages(node.p2pAddress, node.handleMessage, &wg)

	// Wait for all goroutines to finish
	wg.Wait()
}

func (node *GossipNode) handleMessage(msg *GossipMessage) {
	if _, seen := node.messageCache[string(msg.Payload)]; seen {
		return // already seen this message
	}

	// Process message
	node.messageCache[string(msg.Payload)] = struct{}{}
	log.Printf("Received message: %s", string(msg.Payload))

	// Gossip the message to other peers
	node.peersMutex.RLock()
	defer node.peersMutex.RUnlock()
	for peer := range node.peers {
		if err := SendMessage(peer, msg); err != nil {
			log.Printf("Failed to send message to %s: %v", peer, err)
		}
	}
}
