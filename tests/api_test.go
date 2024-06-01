package tests

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/robfig/config"
	"github.com/stretchr/testify/require"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/protocols/api"
	common "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/tests/common"
	"net"
	"testing"
)

func TestAPI(t *testing.T) {
	// start Server on test port
	common.Start(t)

	configFile, readErr := config.ReadDefault("../configs/config.ini")
	require.NoError(t, readErr)

	apiAddress, parseErr := configFile.String("gossip", "api_address_test")
	require.NoError(t, parseErr)

	tests := []struct {
		name        string
		messageType uint16
		message     []byte
	}{
		{
			name:        "AnnounceMsg",
			messageType: api.GossipAnnounce,
			message:     createAnnounceMessage(t),
		},
		{
			name:        "NotifyMsg",
			messageType: api.GossipNotify,
			message:     createNotifyMessage(t),
		},
		{
			name:        "NotificationMsg",
			messageType: api.GossipNotification,
			message:     createNotificationMessage(t),
		},
		{
			name:        "ValidationMsg",
			messageType: api.GossipValidation,
			message:     createValidationMessage(t),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var writeBuffer bytes.Buffer

			err := binary.Write(&writeBuffer, binary.BigEndian, uint16(len(tt.message)+4))
			require.NoError(t, err)

			err = binary.Write(&writeBuffer, binary.BigEndian, tt.messageType)
			require.NoError(t, err)

			err = binary.Write(&writeBuffer, binary.BigEndian, tt.message)
			require.NoError(t, err)

			conn, err := net.Dial("tcp", apiAddress)
			require.NoError(t, err)

			defer func(conn net.Conn) {
				err = conn.Close()
				require.NoError(t, err)
			}(conn)

			write, err := conn.Write(writeBuffer.Bytes())
			require.NoError(t, err)
			fmt.Printf("FROM CLIENT: Sent %d bytes for %s.\n", write, tt.name)

			var readBuffer []byte
			read, err := conn.Read(readBuffer)
			require.NoError(t, err)
			fmt.Printf("FROM CLIENT: Received %d bytes for %s.\n", read, tt.name)
			fmt.Println(string(readBuffer[:read]))
		})
	}
}

func createAnnounceMessage(t *testing.T) []byte {
	const (
		TTL         = uint8(1)
		RESERVED    = uint8(1)
		DATATYPE    = uint16(1)
		MessageData = "Calling announce"
	)

	var buffer bytes.Buffer

	err := binary.Write(&buffer, binary.BigEndian, TTL)
	require.NoError(t, err)

	err = binary.Write(&buffer, binary.BigEndian, RESERVED)
	require.NoError(t, err)

	err = binary.Write(&buffer, binary.BigEndian, DATATYPE)
	require.NoError(t, err)

	err = binary.Write(&buffer, binary.BigEndian, []byte(MessageData))
	require.NoError(t, err)

	return buffer.Bytes()
}

func createNotifyMessage(t *testing.T) []byte {
	const (
		RESERVED = uint16(1)
		DATATYPE = uint16(2)
	)

	var buffer bytes.Buffer

	err := binary.Write(&buffer, binary.BigEndian, RESERVED)
	require.NoError(t, err)

	err = binary.Write(&buffer, binary.BigEndian, DATATYPE)
	require.NoError(t, err)

	return buffer.Bytes()
}

func createNotificationMessage(t *testing.T) []byte {
	const (
		MessageID   = uint16(1)
		DATATYPE    = uint16(3)
		MessageData = "Notification data"
	)

	var buffer bytes.Buffer

	err := binary.Write(&buffer, binary.BigEndian, MessageID)
	require.NoError(t, err)

	err = binary.Write(&buffer, binary.BigEndian, DATATYPE)
	require.NoError(t, err)

	err = binary.Write(&buffer, binary.BigEndian, []byte(MessageData))
	require.NoError(t, err)

	return buffer.Bytes()
}

func createValidationMessage(t *testing.T) []byte {
	const (
		MessageID = uint16(2)
		RESERVED  = uint16(2)
	)

	var buffer bytes.Buffer

	err := binary.Write(&buffer, binary.BigEndian, MessageID)
	require.NoError(t, err)

	err = binary.Write(&buffer, binary.BigEndian, RESERVED)
	require.NoError(t, err)

	return buffer.Bytes()
}
