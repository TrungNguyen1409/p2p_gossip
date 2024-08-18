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

// Mapping of string to Datatype enum
var stringToDatatype = map[string]enum.Datatype{
	"1": enum.Info,
	"2": enum.Warning,
	"3": enum.Error,
}

// Converts string to Datatype enum
func stringToEnum(s string) (enum.Datatype, bool) {
	fmt.Printf("Checking for string: %s\n", s) // Debugging
	datatype, exists := stringToDatatype[s]
	if exists {
		fmt.Printf("Found mapping: %s -> %v\n", s, datatype) // Debugging
	} else {
		fmt.Printf("No mapping found for: %s\n", s) // Debugging
	}
	return datatype, exists
}

// Check checks whether datatype exist
func (am *DatatypeMapper) Check(datatypeStr string) bool {
	// Print the input datatypeStr
	fmt.Println("In Check(), datatypeStr: ", datatypeStr)

	// Print the current state of am.data map
	fmt.Println("Current state of am.data map:", am.data)

	// Attempt to convert string to enum
	datatype, exists := stringToEnum(datatypeStr)
	if !exists {
		fmt.Printf("Datatype '%s' does not exist in enum mapping.\n", datatypeStr)
		fmt.Printf("Datatype value returned: %+v\n", datatype)
		return false
	}

	// Lock the map for reading
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Iterate through the map and check for existence
	for addr, datatypes := range am.data {
		fmt.Printf("Checking address: %s with datatypes: %v\n", addr.String(), datatypes)
		if _, exists := datatypes[datatype]; exists {
			fmt.Printf("Datatype '%s' exists in the map for address %s.\n", datatypeStr, addr.String())
			return true
		}
	}

	// If the datatype was not found
	fmt.Printf("Datatype '%s' does not exist in any address in the map.\n", datatypeStr)
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
