package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
)

type Handler struct {
	validMessageTypes map[uint16]bool
	logger            *logging.Logger
}

func NewHandler(logger *logging.Logger) *Handler {
	return &Handler{logger: logger}
}

// Handle handles incoming connections and dispatches messages based on type
func (handler *Handler) Handle(conn net.Conn, channel chan AnnounceMsg) {
	defer func() {
		if err := conn.Close(); err != nil {
			handler.logger.ErrorF("Error closing connection: %v", err)
		}
		handler.logger.Info("Connection closed")
	}()

	handler.logger.Info("Open connection")

	messageBuffer := new(bytes.Buffer)

	readSize, err := messageBuffer.ReadFrom(conn)
	if err != nil {
		handler.logger.ErrorF("Error reading from connection: %v\n", err)
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
		handler.logger.ErrorF("Error reading size: %v\n", err)
		return
	}

	if int64(size) != readSize {
		sendResponse(conn, "Wrong message size.\n")
		return
	}

	// Read the message type
	if err = binary.Read(reader, binary.BigEndian, &messageType); err != nil {
		handler.logger.ErrorF("Error reading message type: %v\n", err)
		return
	}

	// Handle message based on type
	switch messageType {
	case GossipAnnounce:
		if err = handler.announceHandler(conn, reader, channel); err != nil {
			handler.logger.ErrorF("Error handling ANNOUNCE message: %v\n", err)
		}
	case GossipNotify:
		if err = handler.notifyHandler(conn, reader); err != nil {
			log.Printf("Error handling NOTIFY message: %v\n", err)
		}
	case GossipNotification:
		if err = handler.notificationHandler(conn, reader); err != nil {
			log.Printf("Error handling NOTIFICATION message: %v\n", err)
		}
	case GossipValidation:
		if err = handler.validationHandler(conn, reader); err != nil {
			log.Printf("Error handling VALIDATION message: %v\n", err)
		}
	default:
		sendResponse(conn, fmt.Sprintf("Unknown message type %d.\n", messageType))
	}
}

// announceHandler handles AnnounceMsg
func (handler *Handler) announceHandler(conn net.Conn, reader *bytes.Reader, channel chan AnnounceMsg) error {
	var msg AnnounceMsg
	if err := handler.unmarshallAnnounce(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal announce message: %w", err)
	}

	select {
	// Successfully sent the message
	case channel <- msg:
	default:
		return fmt.Errorf("no receiver available")
	}

	return nil
}

// notifyHandler handles NotifyMsg
func (handler *Handler) notifyHandler(conn net.Conn, reader *bytes.Reader) error {
	var msg NotifyMsg
	if err := handler.unmarshallNotify(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal notify message: %w", err)
	}

	return nil
}

// NotificationHandler handles NotificationMsg
func (handler *Handler) notificationHandler(conn net.Conn, reader *bytes.Reader) error {
	var msg NotificationMsg
	if err := handler.unmarshallNotification(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal notification message: %w", err)
	}

	return nil
}

// ValidationHandler handles ValidationMsg
func (handler *Handler) validationHandler(conn net.Conn, reader *bytes.Reader) error {
	var msg ValidationMsg
	if err := handler.unmarshallValidation(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal validation message: %w", err)
	}

	return nil
}

func (handler *Handler) sendResponse(conn net.Conn, s string) {
	response := []byte(s)
	write, err := conn.Write(response)
	if err != nil {
		handler.logger.ErrorF("Error writing response:", err)
	}
	handler.logger.ErrorF("Sent %d bytes. Message: %s", write, s)
}
