package main

import (
	"encoding/json"
	"fmt"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
)

type Bootstrapper struct {
	mu                  sync.RWMutex
	peersTimeoutList    map[string]time.Time
	seedNodes           []string
	seedNodeLimit       int
	timeout             time.Duration
	cleanupListInterval time.Duration
}

func NewBootstrapper() *Bootstrapper {
	return &Bootstrapper{
		peersTimeoutList:    make(map[string]time.Time),
		seedNodes:           []string{},
		seedNodeLimit:       enum.SeedNodeLimit,
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
	if peer == "" {
		http.Error(w, "Missing peer", http.StatusBadRequest)
		return
	}

	b.mu.Lock()

	// Primitive logic to add peer to seednode list
	if len(b.seedNodes) < b.seedNodeLimit && !contains(b.seedNodes, peer) {
		fmt.Println("register as seed")
		b.seedNodes = append(b.seedNodes, peer)
	}

	b.peersTimeoutList[peer] = time.Now()
	b.mu.Unlock()
	w.WriteHeader(http.StatusOK)
	logger.InfoF("Peer %s registered successfully", peer)

	b.printRegisteredPeers()
	b.printSeedNodes()
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
	rand.Shuffle(len(peers), func(i, j int) { peers[i], peers[j] = peers[j], peers[i] })

	subsetSize := 5 // Limit the number of peers returned to a subset, for example, 5 peer
	if len(peers) < subsetSize {
		subsetSize = len(peers)
	}

	requiredSeedCount := len(b.seedNodes)
	if requiredSeedCount == 0 && len(b.seedNodes) > 0 {
		requiredSeedCount = 1 // Ensure at least one seed node is added
	}

	partialPeers := make([]string, 0, subsetSize)
	seedPeers := make([]string, 0, len(b.seedNodes)) // Separate seedNodes list for response

	seedCount := 0

	for _, seedNode := range b.seedNodes {
		if seedCount < requiredSeedCount && len(partialPeers) < subsetSize {
			partialPeers = append(partialPeers, seedNode)
			seedCount++
		}
		seedPeers = append(seedPeers, seedNode)

	}

	// Fill remaining partialPeers with non-seed nodes
	for _, peer := range peers {
		if !contains(partialPeers, peer) && len(partialPeers) < subsetSize {
			partialPeers = append(partialPeers, peer)
		}
	}

	response := map[string][]string{
		"partialPeers": partialPeers,
		"seedNodes":    seedPeers,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Unable to encode peers", http.StatusInternalServerError)
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
	if _, exists := b.peersTimeoutList[peer]; exists {
		b.peersTimeoutList[peer] = time.Now()
		w.WriteHeader(http.StatusOK)
	} else {
		logger.Error("Peer is not registered with Bootstrapping Server")

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

					// Remove peer from also seedNodes
					for i, seed := range b.seedNodes {
						if seed == peer {
							b.seedNodes = append(b.seedNodes[:i], b.seedNodes[i+1:]...)
							break
						}
					}
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

	counter := 1
	for peer := range b.peersTimeoutList {
		logger.InfoF("%d: %s", counter, peer)
		counter++
	}
}

func (b *Bootstrapper) printSeedNodes() {
	b.mu.RLock()
	defer b.mu.RUnlock()

	logger := logging.NewCustomLogger()
	logger.Info("Current list of seed nodes:")

	for i, peer := range b.seedNodes {
		logger.InfoF("%d: %s", i+1, peer)
	}
}

func main() {
	bootstrapper := NewBootstrapper()
	logger := logging.NewCustomLogger()

	go bootstrapper.RemoveInactivePeers()

	http.HandleFunc("/register", bootstrapper.RegisterPeer)
	http.HandleFunc("/deregister", bootstrapper.DeregisterPeer)
	http.HandleFunc("/peers", bootstrapper.GetPeers)
	http.HandleFunc("/heartbeat", bootstrapper.HandleHeartbeat)

	logger.Info("Bootstrapper server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func contains(peers []string, peer string) bool {
	for _, p := range peers {
		if p == peer {
			return true
		}
	}
	return false
}
