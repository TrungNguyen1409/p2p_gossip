package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
)

type Bootstrapper struct {
	mu      sync.RWMutex
	peers   map[string]time.Time
	timeout time.Duration
}

func NewBootstrapper() *Bootstrapper {
	return &Bootstrapper{
		peers:   make(map[string]time.Time),
		timeout: 1 * time.Second, // Set a timeout for node inactivity
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
	b.peers[peer] = time.Now() // Set the last seen time to now
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
	delete(b.peers, peer)
	b.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (b *Bootstrapper) GetPeers(w http.ResponseWriter, r *http.Request) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	peers := make([]string, 0, len(b.peers))
	for peer := range b.peers {
		peers = append(peers, peer)
	}

	err := json.NewEncoder(w).Encode(peers)
	if err != nil {
		return
	}
}

func (b *Bootstrapper) HandleHeartbeat(w http.ResponseWriter, r *http.Request) {
	peer := r.URL.Query().Get("peer")
	if peer == "" {
		http.Error(w, "Missing peer", http.StatusBadRequest)
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	fmt.Println(peer)
	if _, exists := b.peers[peer]; exists {
		b.peers[peer] = time.Now() // Update the last seen time
		logger := logging.NewCustomLogger()
		logger.DebugF("Received heartbeat from: %s", peer)
		w.WriteHeader(http.StatusOK)
	} else {
		fmt.Println("peer not registered")

		http.Error(w, "Peer not registered", http.StatusBadRequest)
	}
}

func (b *Bootstrapper) RemoveInactivePeers() {
	fmt.Println("activating remove inactive peer")
	// time to start filtering inactive peers
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			for peer, lastSeen := range b.peers {
				if time.Since(lastSeen) > b.timeout {
					logger := logging.NewCustomLogger()
					logger.InfoF("Removing inactive peer: %s", peer)
					delete(b.peers, peer)
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
	for peer := range b.peers {
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
