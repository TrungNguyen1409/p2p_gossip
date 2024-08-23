package enum

const (
	GossipAnnounce     uint16 = 500
	GossipNotify       uint16 = 501
	GossipNotification uint16 = 502
	GossipValidation   uint16 = 503

	PeerAnnounce     uint16 = 511
	PeerLeave        uint16 = 512
	PeerListRequest  uint16 = 513
	PeerListResponse uint16 = 514
)

// Datatype is used to identify the application data Gossip spreads in the network.
type Datatype uint16

// AnnounceMsg represents the structure for GOSSIP ANNOUNCE message
type AnnounceMsg struct {
	TTL      uint8    `json:"ttl"`
	Reserved uint8    `json:"reserved"`
	DataType Datatype `json:"data_type"`
	Data     string   `json:"data"`
}

// NotifyMsg represents the structure for GOSSIP NOTIFY message
type NotifyMsg struct {
	Reserved uint16   `json:"reserved"`
	DataType Datatype `json:"data_type"`
}

// NotificationMsg represents the structure for GOSSIP NOTIFICATION message
type NotificationMsg struct {
	// msgID or uint16
	MessageID string   `json:"message_id"`
	DataType  Datatype `json:"data_type"`
	Data      string   `json:"data"`
}

// ValidationMsg represents the structure for GOSSIP VALIDATION message
type ValidationMsg struct {
	MessageID uint16 `json:"message_id"`
	Reserved  uint16 `json:"reserved"`
}
