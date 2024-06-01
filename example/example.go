package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/robfig/config"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/protocols/api"
	"net"
)

func main() {
	configFile, _ := config.ReadDefault("configs/config.ini")

	apiAddress, _ := configFile.String("gossip", "api_address")

	tests := []struct {
		name        string
		messageType uint16
		message     []byte
	}{
		{
			name:        "AnnounceMsg",
			messageType: api.GossipAnnounce,
			message:     createAnnounceMessage(),
		},
		{
			name:        "NotifyMsg",
			messageType: api.GossipNotify,
			message:     createNotifyMessage(),
		},
		{
			name:        "NotificationMsg",
			messageType: api.GossipNotification,
			message:     createNotificationMessage(),
		},
		{
			name:        "ValidationMsg",
			messageType: api.GossipValidation,
			message:     createValidationMessage(),
		},
	}

	for _, tt := range tests {
		var writeBuffer bytes.Buffer

		_ = binary.Write(&writeBuffer, binary.BigEndian, uint16(len(tt.message)+4))

		_ = binary.Write(&writeBuffer, binary.BigEndian, tt.messageType)

		_ = binary.Write(&writeBuffer, binary.BigEndian, tt.message)

		conn, _ := net.Dial("tcp", apiAddress)

		write, _ := conn.Write(writeBuffer.Bytes())
		fmt.Printf("FROM CLIENT: Sent %d bytes for %s.\n", write, tt.name)

		var msgBytes []byte

		/*_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))*/

		// Read tokens delimited by newline
		msg, err := conn.Read(msgBytes)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("FROM CLIENT: Received %d bytes.\n", msg)
	}
}

func createAnnounceMessage() []byte {
	const (
		TTL         = uint8(1)
		RESERVED    = uint8(1)
		DATATYPE    = uint16(1)
		MessageData = "Calling announce"
	)

	var buffer bytes.Buffer

	err := binary.Write(&buffer, binary.BigEndian, TTL)

	err = binary.Write(&buffer, binary.BigEndian, RESERVED)

	err = binary.Write(&buffer, binary.BigEndian, DATATYPE)

	err = binary.Write(&buffer, binary.BigEndian, []byte(MessageData))

	if err != nil {
		panic(err)
	}

	return buffer.Bytes()
}

func createNotifyMessage() []byte {
	const (
		RESERVED = uint16(1)
		DATATYPE = uint16(2)
	)

	var buffer bytes.Buffer

	err := binary.Write(&buffer, binary.BigEndian, RESERVED)

	err = binary.Write(&buffer, binary.BigEndian, DATATYPE)

	if err != nil {
		panic(err)
	}

	return buffer.Bytes()
}

func createNotificationMessage() []byte {
	const (
		MessageID   = uint16(1)
		DATATYPE    = uint16(3)
		MessageData = "Notification data"
	)

	var buffer bytes.Buffer

	err := binary.Write(&buffer, binary.BigEndian, MessageID)

	err = binary.Write(&buffer, binary.BigEndian, DATATYPE)

	err = binary.Write(&buffer, binary.BigEndian, []byte(MessageData))

	if err != nil {
		panic(err)
	}

	return buffer.Bytes()
}

func createValidationMessage() []byte {
	const (
		MessageID = uint16(2)
		RESERVED  = uint16(2)
	)

	var buffer bytes.Buffer

	err := binary.Write(&buffer, binary.BigEndian, MessageID)

	err = binary.Write(&buffer, binary.BigEndian, RESERVED)

	if err != nil {
		panic(err)
	}

	return buffer.Bytes()
}
