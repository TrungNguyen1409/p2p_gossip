package main

import (
	"encoding/json"
	"fmt"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"log"
	"net/http"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
)

type Bootstrapper struct {
	mu                  sync.RWMutex
	peersTimeoutList    map[string]time.Time
	timeout             time.Duration
	cleanupListInterval time.Duration
}

func NewBootstrapper() *Bootstrapper {
	return &Bootstrapper{
		peersTimeoutList:    make(map[string]time.Time),
		timeout:             enum.Timeout,
		cleanupListInterval: enum.CleanupListInterval, // Set a timeout for node inactivity
	}
}

func (b *Bootstrapper) RegisterPeer(w http.ResponseWriter, r *http.Request) {
	logger := logging.NewCustomLogger()
	logger.Debug("Peer starts Registering...")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	peer := r.FormValue("peer")
	logger.DebugF("peer: %s", peer)

	if peer == "" {
		http.Error(w, "Missing peer", http.StatusBadRequest)
		return
	}

	b.mu.Lock()
	b.peersTimeoutList[peer] = time.Now() // Set the last seen time to now
	b.mu.Unlock()
	logger.Info("Registering successful")
	w.WriteHeader(http.StatusOK)

	b.printRegisteredPeers()
}

func (b *Bootstrapper) DeregisterPeer(w http.ResponseWriter, r *http.Request) {
	peer := r.URL.Query().Get("peer")
	if peer == "" {
		http.Error(w, "Missing peer", http.StatusBadRequest)
		return
	}

	b.mu.Lock()
	delete(b.peersTimeoutList, peer)
	b.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (b *Bootstrapper) GetPeers(w http.ResponseWriter, r *http.Request) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	peers := make([]string, 0, len(b.peersTimeoutList))
	for peer := range b.peersTimeoutList {
		peers = append(peers, peer)
	}

	err := json.NewEncoder(w).Encode(peers)
	if err != nil {
		return
	}
}

func (b *Bootstrapper) HandleHeartbeat(w http.ResponseWriter, r *http.Request) {
	logger := logging.NewCustomLogger()

	peer := r.URL.Query().Get("peer")
	if peer == "" {
		http.Error(w, "Missing peer", http.StatusBadRequest)
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	fmt.Println(peer)
	if _, exists := b.peersTimeoutList[peer]; exists {
		b.peersTimeoutList[peer] = time.Now() // Update the last seen time
		logger.DebugF("Received heartbeat from: %s", peer)
		w.WriteHeader(http.StatusOK)
	} else {
		fmt.Println("peer not registered")

		http.Error(w, "Peer not registered", http.StatusBadRequest)
	}
}

func (b *Bootstrapper) RemoveInactivePeers() {
	logger := logging.NewCustomLogger()

	ticker := time.NewTicker(b.cleanupListInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			for peer, lastSeen := range b.peersTimeoutList {

				if time.Since(lastSeen) > b.timeout {
					logger.InfoF("Removing inactive peer: %s", peer)
					delete(b.peersTimeoutList, peer)
				}
			}
			b.mu.Unlock()
		}
	}
}

func (b *Bootstrapper) printRegisteredPeers() {
	b.mu.RLock()
	defer b.mu.RUnlock()

	logger := logging.NewCustomLogger()
	logger.Info("Current list of registered peers:")
	for peer := range b.peersTimeoutList {
		fmt.Println(peer)
	}
}

func main() {
	bootstrapper := NewBootstrapper()

	go bootstrapper.RemoveInactivePeers()

	http.HandleFunc("/register", bootstrapper.RegisterPeer)
	http.HandleFunc("/deregister", bootstrapper.DeregisterPeer)
	http.HandleFunc("/peers", bootstrapper.GetPeers)
	http.HandleFunc("/heartbeat", bootstrapper.HandleHeartbeat)

	fmt.Println("Bootstrapper server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
