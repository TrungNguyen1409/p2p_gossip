package common

import (
	"fmt"
	"net"
	"sync"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
)

// DatatypeMapper map address -> another map value: enum.Datatype -> boolean, that indicates presence of that type
type DatatypeMapper struct {
	mu   sync.RWMutex
	data map[net.Addr]map[enum.Datatype]bool
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

func (am *DatatypeMapper) CheckNotify(datatype int) bool {
	// Check if the datatype is GossipNotify
	if datatype != int(enum.GossipNotification) {
		fmt.Printf("Datatype '%d' is not of type GossipNotification.\n", datatype)
		return false
	}

	fmt.Println("Current state of am.data map:", am.data)

	notifyMsgType := enum.Datatype(enum.GossipNotify)

	am.mu.RLock()
	defer am.mu.RUnlock()

	// Iterate through the map and check for existence of Notify demand
	for addr, datatypes := range am.data {
		fmt.Printf("Checking address: %s with datatypes: %v\n", addr.String(), datatypes)
		if _, exists := datatypes[notifyMsgType]; exists {
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
