package enum

import "time"

/*
•	Timeout: The maximum time a node is allowed to go without sending a heartbeat message to the bootstrapper. After this period, the node is considered inactive.
•	CleanupListInterval: The time interval at which the bootstrapper removes inactive nodes from its list of active nodes.
•	PeriodicBootstrapTicker: The time interval in which nodes periodically fetch the updated peer list from the bootstrapper.
•	HeartbeatTicker: The interval at which a node sends a heartbeat message to the bootstrapper to signal that it is still active.
•	GossipInterval: The time interval at which nodes send gossip messages to other nodes in the network to spread information.
•	SeedNodeLimit: Maximum number of SeedNode in the network
*/
/*
const (
	Timeout                 = 120 * time.Second
	CleanupListInterval     = 120 * time.Second
	PeriodicBootstrapTicker = 60 * time.Second
	HeartbeatTicker         = 20 * time.Second
	GossipInterval          = 60 * time.Second
	SeedNodeLimit           = 3
)*/

const (
	Timeout                 = 240 * time.Second
	CleanupListInterval     = 240 * time.Second
	PeriodicBootstrapTicker = 120 * time.Second
	HeartbeatTicker         = 60 * time.Second
	GossipInterval          = 60 * time.Second
	SeedNodeLimit           = 3
)
