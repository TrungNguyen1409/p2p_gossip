package api

// AnnounceMsg represents the structure for GOSSIP ANNOUNCE message
type AnnounceMsg struct {
	TTL      uint8  `json:"ttl"`
	Reserved uint8  `json:"reserved"`
	DataType uint16 `json:"data_type"`
	Data     string `json:"data"`
}

// NotifyMsg represents the structure for GOSSIP NOTIFY message
type NotifyMsg struct {
	Reserved uint16 `json:"reserved"`
	DataType uint16 `json:"data_type"`
}

// NotificationMsg represents the structure for GOSSIP NOTIFICATION message
type NotificationMsg struct {
	MessageID uint16 `json:"message_id"`
	DataType  uint16 `json:"data_type"`
	Data      string `json:"data"`
}

// ValidationMsg represents the structure for GOSSIP VALIDATION message
type ValidationMsg struct {
	MessageID uint16 `json:"message_id"`
	Reserved  uint16 `json:"reserved"`
}
