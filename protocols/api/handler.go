package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/common"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
)

type Handler struct {
	conn                net.Conn
	logger              *logging.Logger
	announceMsgChan     chan enum.AnnounceMsg
	notificationMsgChan chan enum.NotificationMsg
	datatypeMapper      *common.DatatypeMapper
}

func NewHandler(conn net.Conn, logger *logging.Logger, announceMsgChan chan enum.AnnounceMsg, notificationMsgChan chan enum.NotificationMsg, datatypeMapper *common.DatatypeMapper) *Handler {
	return &Handler{conn: conn, logger: logger, announceMsgChan: announceMsgChan, notificationMsgChan: notificationMsgChan, datatypeMapper: datatypeMapper}
}

// Handle handles incoming connections and dispatches messages based on type
func (h *Handler) Handle() {
	h.logger.InfoF("Open connection with %s\n", h.conn.RemoteAddr())

	messageBuffer := make([]byte, 1024)
	readSize, err := h.conn.Read(messageBuffer)

	h.logger.DebugF("Reading from %s\n", h.conn.RemoteAddr())

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

	reader := bytes.NewReader(messageBuffer)

	var size uint16
	var messageType uint16

	// Read the size
	if err = binary.Read(reader, binary.BigEndian, &size); err != nil {
		h.logger.ErrorF("Error reading size: %v\n", err)
		return
	}

	if int(size) != readSize {
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
	case enum.GossipNotify:
		if err = h.notifyHandler(reader); err != nil {
			h.logger.ErrorF("Error handling NOTIFY message: %v\n", err)
		}
		h.datatypeMapper.Print()
	case enum.GossipAnnounce:
		if err = h.announceHandler(reader); err != nil {
			h.logger.ErrorF("Error handling ANNOUNCE message: %v\n", err)
		}
	}
}

// announceHandler handles AnnounceMsg
func (h *Handler) announceHandler(reader *bytes.Reader) error {
	defer func() {
		if err := h.conn.Close(); err != nil {
			h.logger.ErrorF("Error closing connection: %v", err)
		}
		h.logger.Info("Connection closed")
	}()

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
	defer func() {
		if err := h.conn.Close(); err != nil {
			h.logger.ErrorF("Error closing connection: %v", err)
		}
		h.logger.Info("Connection closed")
	}()

	var msg enum.NotifyMsg
	if err := h.unmarshallNotify(reader, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal notify message: %w", err)
	}

	h.datatypeMapper.Add(h.conn.RemoteAddr(), msg.DataType)

	for {
		// Wait for a message to be received on the channel
		notificationMsg, ok := <-h.notificationMsgChan
		if !ok {
			// If the channel is closed, exit the loop
			return fmt.Errorf("notification channel closed")
		} else {
			h.logger.Info("Got notification message")
			h.logger.InfoF("%d and %d \n", notificationMsg.DataType, msg.DataType)
			if notificationMsg.DataType == msg.DataType {
				sendNotificationMessage(h.conn, notificationMsg, h.logger)
			}
		}
	}
}

func (h *Handler) sendResponse(s string) {
	response := []byte(s)
	write, err := h.conn.Write(response)
	if err != nil {
		h.logger.ErrorF("Error writing response:", err)
	}
	h.logger.ErrorF("Sent %d bytes. Message: %s", write, s)
}
