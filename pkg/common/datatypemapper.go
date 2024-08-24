package common

import (
	"fmt"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
	"net"
	"sync"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
)

// DatatypeMapper map address -> another map value: enum.Datatype -> boolean, that indicates presence of that type
type DatatypeMapper struct {
	mu     sync.RWMutex
	data   map[net.Addr]map[enum.Datatype]bool
	logger *logging.Logger
}

// NewMap initializes a new DatatypeMapper.
func NewMap() *DatatypeMapper {
	return &DatatypeMapper{
		data: make(map[net.Addr]map[enum.Datatype]bool),
	}
}

// Add adds a new datatype to a specific address in the DatatypeMapper.
func (am *DatatypeMapper) Add(addr net.Addr, datatype enum.Datatype) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if _, exists := am.data[addr]; !exists {
		am.data[addr] = make(map[enum.Datatype]bool)
	}
	am.data[addr][datatype] = true
}

func (am *DatatypeMapper) CheckNotify(datatype enum.Datatype) bool {

	am.mu.RLock()
	defer am.mu.RUnlock()

	// Iterate through the map and check for existence of Notify demand
	for addr, datatypes := range am.data {
		fmt.Printf("Checking address: %s with datatypes: %v\n", addr.String(), datatypes)
		if _, exists := datatypes[datatype]; exists {
			fmt.Printf("Datatype '%d' (GossipNotify) exists in the map for address %s.\n", datatype, addr.String())
			return true
		}
	}

	fmt.Printf("Datatype '%d' (GossipNotify) does not exist in any address in the map.\n", datatype)
	return false
}

// Print displays the current state of the DatatypeMapper.
func (am *DatatypeMapper) Print() {
	am.mu.RLock()
	defer am.mu.RUnlock()
	for addr, datatypes := range am.data {
		fmt.Printf("Address: %s\n", addr.String())
		for datatype := range datatypes {
			fmt.Printf("- Datatype: %d\n", int(datatype))
		}
	}
}

func (am *DatatypeMapper) GetAddressesByType(datatype enum.Datatype) []net.Addr {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var addresses []net.Addr

	// Iterate through the map and collect addresses with the specified datatype
	for addr, datatypes := range am.data {
		if exists := datatypes[datatype]; exists {
			addresses = append(addresses, addr)
		}
	}

	return addresses
}
