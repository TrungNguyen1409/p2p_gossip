package p2p

import (
	"encoding/json"
	"fmt"
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

	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(peerList), func(i, j int) { peerList[i], peerList[j] = peerList[j], peerList[i] })

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

func (node *GossipNode) periodicGossip() {
	ticker := time.NewTicker(node.gossipInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			node.peersMutex.RLock()
			peers := make([]string, 0, len(node.peers))
			for peer := range node.peers {
				peers = append(peers, peer)
			}
			node.peersMutex.RUnlock()

			/*for _, peer := range peers {
				go node.sendPeerList(peer, logger)
			}*/
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
