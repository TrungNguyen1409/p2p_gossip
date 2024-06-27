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
