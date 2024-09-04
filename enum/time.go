package enum

import "time"

const (
	Timeout                 = 45 * time.Second
	CleanupListInterval     = 60 * time.Second
	PeriodicBootstrapTicker = 15 * time.Second
	HeartbeatTicker         = 15 * time.Second
	GossipInterval          = 60 * time.Second
)
