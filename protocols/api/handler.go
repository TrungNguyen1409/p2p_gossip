package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"log"
	"net"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/common"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
)

type Handler struct {
	conn            net.Conn
	logger          *logging.Logger
	announceMsgChan chan enum.AnnounceMsg
	datatypeMapper  *common.DatatypeMapper
}

func NewHandler(conn net.Conn, logger *logging.Logger, announceMsgChan chan enum.AnnounceMsg, datatypeMapper *common.DatatypeMapper) *Handler {
	return &Handler{conn: conn, logger: logger, announceMsgChan: announceMsgChan, datatypeMapper: datatypeMapper}
}

// Handle handles incoming connections and dispatches messages based on type
func (h *Handler) Handle() {
	defer func() {
		if err := h.conn.Close(); err != nil {
			h.logger.ErrorF("Error closing connection: %v", err)
		}
		h.logger.Info("Connection closed")
	}()

	h.conn.RemoteAddr()

	h.logger.Info("Open connection")

	messageBuffer := new(bytes.Buffer)

	readSize, err := messageBuffer.ReadFrom(h.conn)
	if err != nil {
		h.logger.ErrorF("Error reading from connection: %v\n", err)
		return
	}

	if readSize == 0 {
		h.sendResponse("Empty message.\n")
		return
	} else if readSize < 4 {
		h.sendResponse("Message too short.\n")
		return
	}

	reader := bytes.NewReader(messageBuffer.Bytes())

	var size uint16
	var messageType uint16

	// Read the size
	if err = binary.Read(reader, binary.BigEndian, &size); err != nil {
		h.logger.ErrorF("Error reading size: %v\n", err)
		return
	}

	if int64(size) != readSize {
		h.sendResponse("Wrong message size.\n")
		return
	}

	// Read the message type
	if err = binary.Read(reader, binary.BigEndian, &messageType); err != nil {
		h.logger.ErrorF("Error reading message type: %v\n", err)
		return
	}

	// Handle message based on type
	switch messageType {
	case enum.GossipAnnounce:
		if err = h.announceHandler(reader); err != nil {
			h.logger.ErrorF("Error handling ANNOUNCE message: %v\n", err)
		}
	case enum.GossipNotify:
		if err = h.notifyHandler(reader); err != nil {
			log.Printf("Error handling NOTIFY message: %v\n", err)
		}
	case enum.GossipNotification:
		if err = h.notificationHandler(reader); err != nil {
			log.Printf("Error handling NOTIFICATION message: %v\n", err)
		}
	case enum.GossipValidation:
		if err = h.validationHandler(reader); err != nil {
			log.Printf("Error handling VALIDATION message: %v\n", err)
		}
	default:
		h.sendResponse(fmt.Sprintf("Unknown message type %d.\n", messageType))
	}

	h.datatypeMapper.Print()
}

// announceHandler handles AnnounceMsg
func (h *Handler) announceHandler(reader *bytes.Reader) error {
	var msg enum.AnnounceMsg
	if err := h.unmarshallAnnounce(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal announce message: %w", err)
	}

	select {
	// Successfully sent the message
	case h.announceMsgChan <- msg:
	default:
		return fmt.Errorf("no receiver available")
	}

	return nil
}

// notifyHandler handles NotifyMsg
func (h *Handler) notifyHandler(reader *bytes.Reader) error {
	var msg enum.NotifyMsg
	if err := h.unmarshallNotify(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal notify message: %w", err)
	}

	h.datatypeMapper.Add(h.conn.RemoteAddr(), msg.DataType)

	return nil
}

// NotificationHandler handles NotificationMsg
func (h *Handler) notificationHandler(reader *bytes.Reader) error {
	var msg enum.NotificationMsg
	if err := h.unmarshallNotification(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal notification message: %w", err)
	}

	return nil
}

// ValidationHandler handles ValidationMsg
func (h *Handler) validationHandler(reader *bytes.Reader) error {
	var msg enum.ValidationMsg
	if err := h.unmarshallValidation(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal validation message: %w", err)
	}

	return nil
}

func (h *Handler) sendResponse(s string) {
	response := []byte(s)
	write, err := h.conn.Write(response)
	if err != nil {
		h.logger.ErrorF("Error writing response:", err)
	}
	h.logger.ErrorF("Sent %d bytes. Message: %s", write, s)
}
