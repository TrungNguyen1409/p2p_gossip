package p2p

import (
	"encoding/json"
	"fmt"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/pow"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
	pb "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/proto"
)

/* --------------------------------- GOSSIPPING ---------------------------------- */

func (node *GossipNode) gossip(msg *pb.GossipMessage, logger *logging.Logger) {
	pow.CalculateAndAddNonce(msg)

	logger.DebugF(string(msg.Ttl))
	if msg.Ttl < 1 {
		logger.Info("Message TTL expired, not gossiping.")
		return
	}

	node.peersMutex.RLock()
	defer node.peersMutex.RUnlock()

	peerList := make([]string, 0, len(node.peers))
	for peer := range node.peers {
		peerList = append(peerList, peer)
	}

	//rand.New(rand.NewSource(time.Now().UnixNano()))
	//rand.Shuffle(len(peerList), func(i, j int) { peerList[i], peerList[j] = peerList[j], peerList[i] })

	fanout := node.fanout

	if len(peerList) < fanout {
		fanout = len(peerList)
	}

	for i := 0; i < fanout; i++ {
		peer := peerList[i]
		logger.InfoF("gossiping with: %s", peer)
		go func(peer string) {
			if err := send(peer, msg); err != nil {
				logger.ErrorF("Failed to send message to %s: %v", peer, err)
			} else {
				logger.InfoF("Message sent to %s", peer)
			}
		}(peer)
	}
}

func (node *GossipNode) periodicPeerListRequest() {
	ticker := time.NewTicker(node.gossipInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger := logging.NewCustomLogger()

			node.seedNodesMutex.RLock()
			if len(node.seedNodes) > 0 {
				// Create a list of seed nodes excluding the node's own address
				seedNodeList := make([]string, 0, len(node.seedNodes))
				for seed := range node.seedNodes {
					if seed != node.p2pAddress {
						seedNodeList = append(seedNodeList, seed)
					}
				}

				if len(seedNodeList) > 0 {
					// Select a random peer
					randPeer := seedNodeList[rand.Intn(len(seedNodeList))]
					logger.InfoF("Requesting peer list from: %s", randPeer)

					// Request the peer list from the selected random peer
					node.requestPeerList(randPeer, logger)
				} else {
					logger.Info("No other peers available to request peer list.")
				}
			} else {
				logger.Info("No seed nodes available to request peer list.")
			}
			node.seedNodesMutex.RUnlock()
		}
	}
}
func (node *GossipNode) sendPeerList(peerAddress string, logger *logging.Logger) {
	node.peersMutex.RLock()
	peerList := make([]string, 0, len(node.peers))
	for peer := range node.peers {
		peerList = append(peerList, peer)
	}
	node.peersMutex.RUnlock()

	data, err := json.Marshal(peerList)
	if err != nil {
		logger.ErrorF("Failed to marshal peer list: %v", err)
		return
	}

	resp, err := http.Post(fmt.Sprintf("http://%s/gossip", peerAddress), "application/json", strings.NewReader(string(data)))
	if err != nil || resp.StatusCode != http.StatusOK {
		logger.ErrorF("Failed to send gossip to %s: %v", peerAddress, err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
}

func (node *GossipNode) requestPeerList(targetPeer string, logger *logging.Logger) {
	requestMsg := &pb.GossipMessage{
		MessageId: uint32(generate16BitRandomInteger()),
		From:      node.p2pAddress,
		Type:      int32(enum.PeerListRequest),
		Ttl:       1,
	}
	pow.CalculateAndAddNonce(requestMsg)

	err := send(targetPeer, requestMsg)
	if err != nil {
		logger.ErrorF("Failed to request peer list from %s: %v", targetPeer, err)
	} else {
		logger.InfoF("PeerListRequest sent to %s", targetPeer)
	}
}

func (node *GossipNode) respondWithPeerList(targetPeer string, logger *logging.Logger) {
	node.peersMutex.RLock()
	defer node.peersMutex.RUnlock()

	peers := make([]string, 0, len(node.peers))
	for peer := range node.peers {
		peers = append(peers, peer)
	}

	peerListMsg, err := json.Marshal(peers)
	if err != nil {
		logger.ErrorF("Failed to marshal peer list: %v", err)
		return
	}

	responseMsg := &pb.GossipMessage{
		MessageId: uint32(generate16BitRandomInteger()),
		From:      node.p2pAddress,
		Type:      int32(enum.PeerListResponse),
		Payload:   peerListMsg,
		Ttl:       1,
	}
	pow.CalculateAndAddNonce(responseMsg)

	err = send(targetPeer, responseMsg)
	if err != nil {
		logger.ErrorF("Failed to send peer list to %s: %v", targetPeer, err)
	} else {
		logger.InfoF("PeerListResponse sent to %s", targetPeer)
	}
}

func (node *GossipNode) updateByPeerListResponse(receivedPeers []string, logger *logging.Logger) {
	node.peersMutex.Lock()
	defer node.peersMutex.Unlock()
	for _, peer := range receivedPeers {
		if _, exists := node.peers[peer]; !exists {
			node.peers[peer] = struct{}{}
			logger.InfoF("Added new peer: %s", peer)
		} else {
			logger.InfoF("peer list already contains peer: %s", peer)

		}
	}
	node.removeNodeByExceedDegree(logger)
	node.PrintPeerLists()
}

func (node *GossipNode) removeNodeByExceedDegree(logger *logging.Logger) {
	for len(node.peers) > node.degree {
		for oldestPeer := range node.peers {
			delete(node.peers, oldestPeer)
			logger.InfoF("Degree size exceeded, removed oldest peer: %s", oldestPeer)
			break
		}
	}
}

func (node *GossipNode) announceNewPeer() {
	logger := logging.NewCustomLogger()
	logger.InfoF("Announcing new peer: %s", node.p2pAddress)

	announceMsg := &pb.GossipMessage{
		MessageId: uint32(generate16BitRandomInteger()),
		Payload:   []byte(node.p2pAddress),
		From:      node.p2pAddress,
		Type:      int32(enum.PeerJoinAnnounce),
		Ttl:       int32(5),
	}

	node.gossip(announceMsg, logger)
}

func (node *GossipNode) updateByPeerJoin(peerAddress string, logger *logging.Logger) {
	node.peersMutex.Lock()
	if _, exists := node.peers[peerAddress]; !exists {
		node.peers[peerAddress] = struct{}{}
		logger.InfoF("New peer announced and added: %s", peerAddress)
	} else {
		logger.InfoF("Peer %s already known", peerAddress)
	}

	node.removeNodeByExceedDegree(logger)
	node.peersMutex.Unlock()
}

func (node *GossipNode) announceLeave() {
	logger := logging.NewCustomLogger()
	logger.InfoF("Announcing that peer %s is leaving", node.p2pAddress)

	leaveMsg := &pb.GossipMessage{
		MessageId: uint32(generate16BitRandomInteger()),
		Payload:   []byte(node.p2pAddress),
		From:      node.p2pAddress,
		Type:      int32(enum.PeerLeaveAnnounce),
		Ttl:       int32(5),
	}

	node.gossip(leaveMsg, logger)
}

func (node *GossipNode) updateByPeerLeave(peerAddress string, logger *logging.Logger) {
	node.peersMutex.Lock()
	delete(node.peers, peerAddress)
	node.peersMutex.Unlock()

	logger.InfoF("Peer %s left and removed from peer list", peerAddress)
}

func (node *GossipNode) ShutDown() {
	logger := logging.NewCustomLogger()
	logger.Info("Shutting down gracefully...")

	node.announceLeave()

	logger.Info("Node has shut down successfully.")
}
