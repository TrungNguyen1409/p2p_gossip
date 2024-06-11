package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

// Handler handles incoming connections and dispatches messages based on type
func Handler(conn net.Conn, channel chan AnnounceMsg) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v\n", err)
		}
		fmt.Printf("Connection closed!\n----------------------------------\n")
	}()

	fmt.Printf("\n----------------------------------\nOpen connection with %s\n", conn.RemoteAddr())

	messageBuffer := new(bytes.Buffer)

	readSize, err := messageBuffer.ReadFrom(conn)
	if err != nil {
		log.Printf("Error reading from connection: %v\n", err)
		return
	}

	if readSize == 0 {
		sendResponse(conn, "Empty message.\n")
		return
	} else if readSize < 4 {
		sendResponse(conn, "Message too short.\n")
		return
	}

	reader := bytes.NewReader(messageBuffer.Bytes())

	var size uint16
	var messageType uint16

	// Read the size
	if err = binary.Read(reader, binary.BigEndian, &size); err != nil {
		log.Printf("Error reading size: %v\n", err)
		return
	}

	if int64(size) != readSize {
		sendResponse(conn, "Wrong message size.\n")
		return
	}

	// Read the message type
	if err = binary.Read(reader, binary.BigEndian, &messageType); err != nil {
		log.Printf("Error reading message type: %v\n", err)
		return
	}

	// Handle message based on type
	switch messageType {
	case GossipAnnounce:
		if err = AnnounceHandler(conn, reader, channel); err != nil {
			log.Printf("Error handling ANNOUNCE message: %v\n", err)
		}
	case GossipNotify:
		if err = NotifyHandler(conn, reader); err != nil {
			log.Printf("Error handling NOTIFY message: %v\n", err)
		}
	case GossipNotification:
		if err = NotificationHandler(conn, reader); err != nil {
			log.Printf("Error handling NOTIFICATION message: %v\n", err)
		}
	case GossipValidation:
		if err = ValidationHandler(conn, reader); err != nil {
			log.Printf("Error handling VALIDATION message: %v\n", err)
		}
	default:
		sendResponse(conn, fmt.Sprintf("Unknown message type %d.\n", messageType))
	}
}

// AnnounceHandler handles AnnounceMsg
func AnnounceHandler(conn net.Conn, reader *bytes.Reader, channel chan AnnounceMsg) error {
	var msg AnnounceMsg
	if err := unmarshallAnnounce(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal announce message: %w", err)
	}

	select {
	// Successfully sent the message
	case channel <- msg:
	default:
		return fmt.Errorf("channel is full or no receiver available")
	}

	return nil
}

// NotifyHandler handles NotifyMsg
func NotifyHandler(conn net.Conn, reader *bytes.Reader) error {
	var msg NotifyMsg
	if err := unmarshallNotify(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal notify message: %w", err)
	}

	return nil
}

// NotificationHandler handles NotificationMsg
func NotificationHandler(conn net.Conn, reader *bytes.Reader) error {
	var msg NotificationMsg
	if err := unmarshallNotification(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal notification message: %w", err)
	}

	return nil
}

// ValidationHandler handles ValidationMsg
func ValidationHandler(conn net.Conn, reader *bytes.Reader) error {
	var msg ValidationMsg
	if err := unmarshallValidation(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal validation message: %w", err)
	}

	return nil
}
