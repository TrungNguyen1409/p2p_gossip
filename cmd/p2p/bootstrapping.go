package p2p

import (
	"encoding/json"
	"fmt"
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
	ticker := time.NewTicker(5 * time.Second)
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
func (node *GossipNode) sendHeartbeat() {
	ticker := time.NewTicker(5 * time.Second) // Adjust the interval as needed
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger := logging.NewCustomLogger()

			// Split the p2pAddress into host and port
			host, port, err := net.SplitHostPort(node.p2pAddress)
			if err != nil {
				logger.ErrorF("Invalid p2pAddress format: %v", err)
				continue
			}

			// Check if the host is "localhost" and replace it with "127.0.0.1"
			if host == "localhost" {
				host = "127.0.0.1"
			}

			// Recombine host and port into the final address
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
