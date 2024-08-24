package p2p

import (
	"encoding/json"
	"fmt"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
	"io"
	"net/http"
	"net/url"
	"time"
)

/* --------------------------------- BOOTSTRAPPING ---------------------------------- */

func (node *GossipNode) registerWithBootstrapper(p2pAddress string) error {
	logger := logging.NewCustomLogger()
	resp, err := http.PostForm(node.bootstrapURL+"/register", url.Values{"peer": {p2pAddress}})
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.ErrorF("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logger.ErrorF("failed to register with bootstrapper, status code: %d", resp.StatusCode)
		return nil
	}
	return nil
}

func (node *GossipNode) getInitialPeers() error {

	logger := logging.NewCustomLogger()

	resp, err := http.Get(node.bootstrapURL + "/peers")
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.ErrorF("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get peers from bootstrapper, status code: %d", resp.StatusCode)
	}

	var peers []string
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return err
	}

	node.peersMutex.Lock()
	defer node.peersMutex.Unlock()
	for _, peer := range peers {
		if peer != node.p2pAddress {
			node.peers[peer] = struct{}{}
		}
	}
	logger.InfoF("Initial peers fetched successfully: %v", peers)
	return nil
}

func (node *GossipNode) periodicBootstrapping() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger := logging.NewCustomLogger()
			logger.Info("Fetching new list of peers from bootstrapper...")

			if err := node.getInitialPeers(); err != nil {
				logger.ErrorF("Failed to fetch peers from bootstrapper: %v", err)
			} else {
				logger.Info("Successfully updated peers from bootstrapper.")
			}
		}
	}
}
