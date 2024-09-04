package p2p

import (
	"encoding/json"
	"fmt"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

/* --------------------------------- BOOTSTRAPPING ---------------------------------- */

func (node *GossipNode) registerWithBootstrapper(p2pAddress string) error {
	logger := logging.NewCustomLogger()
	logger.InfoF("Registering with: %v", node.bootstrapURL)

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
	logger.InfoF("Peers list updated successfully from: %v", node.bootstrapURL)
	return nil
}

func (node *GossipNode) periodicBootstrapping() {
	ticker := time.NewTicker(enum.PeriodicBootstrapTicker)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger := logging.NewCustomLogger()

			if err := node.getInitialPeers(); err != nil {
				logger.ErrorF("Failed to fetch peers from bootstrapper: %v", err)
			}
		}
	}
}
func (node *GossipNode) sendHeartbeat() {
	ticker := time.NewTicker(enum.HeartbeatTicker)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger := logging.NewCustomLogger()

			host, port, err := net.SplitHostPort(node.p2pAddress)
			if err != nil {
				logger.ErrorF("Invalid p2pAddress format: %v", err)
				continue
			}

			if host == "localhost" {
				host = "127.0.0.1"
			}

			address := net.JoinHostPort(host, port)

			heartBeatURL := fmt.Sprintf("%s/heartbeat?peer=%s", node.bootstrapURL, address)
			resp, err := http.Get(heartBeatURL)
			if err != nil {
				logger.ErrorF("Failed to send heartbeat to bootstrapper: %v", err)
				continue
			}

			defer func(Body io.ReadCloser) {
				if err := Body.Close(); err != nil {
					logger.ErrorF("Failed to close heartbeat response body: %v", err)
				}
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				logger.ErrorF("Heartbeat response status code not OK: %d", resp.StatusCode)
			} else {
				logger.DebugF("Heartbeat successfully sent to bootstrapper for peer: %s", address)
			}
		}
	}
}
