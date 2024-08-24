package api

import (
	"bytes"
	"encoding/binary"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
)

// unmarshallAnnounce parses the JSON data manually, and populates the AnnounceMsg struct
func (h *Handler) unmarshallAnnounce(readBuffer *bytes.Reader, msg *enum.AnnounceMsg) error {
	if err := binary.Read(readBuffer, binary.BigEndian, &msg.TTL); err != nil {
		h.logger.ErrorF("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.Reserved); err != nil {
		h.logger.ErrorF("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.DataType); err != nil {
		h.logger.ErrorF("Error reading message type:", err)
		return err
	}

	msgBuf := make([]byte, readBuffer.Len())

	if err := binary.Read(readBuffer, binary.BigEndian, &msgBuf); err != nil {
		h.logger.ErrorF("Error reading message type:", err)
		return err
	} else {
		msg.Data = string(msgBuf)
	}

	h.logger.InfoF("Received GOSSIP_ANNOUNCE message %v", *msg)

	return nil
}

// unmarshallNotify parses the JSON data manually, and populates the NotifyMsg struct
func (h *Handler) unmarshallNotify(readBuffer *bytes.Reader, msg *enum.NotifyMsg) error {
	if err := binary.Read(readBuffer, binary.BigEndian, &msg.Reserved); err != nil {
		h.logger.ErrorF("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.DataType); err != nil {
		h.logger.ErrorF("Error reading message type:", err)
		return err
	}

	h.logger.InfoF("Received GOSSIP_NOTIFY message %v", *msg)

	return nil
}

// marshallNotification parses the JSON data manually, and populates the NotificationMsg struct
func (h *Handler) unmarshallNotification(readBuffer *bytes.Reader, msg *enum.NotificationMsg) error {
	if err := binary.Read(readBuffer, binary.BigEndian, &msg.MessageID); err != nil {
		h.logger.ErrorF("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.DataType); err != nil {
		h.logger.ErrorF("Error reading message type:", err)
		return err
	}

	msgBuf := make([]byte, readBuffer.Len())

	if err := binary.Read(readBuffer, binary.BigEndian, &msgBuf); err != nil {
		h.logger.ErrorF("Error reading message type:", err)
		return err
	} else {
		msg.Data = string(msgBuf)
	}

	h.logger.InfoF("Received message %v", *msg)

	return nil
}

// unmarshallValidation parses the JSON data manually, and populates the ValidationMsg struct
func (h *Handler) unmarshallValidation(readBuffer *bytes.Reader, msg *enum.ValidationMsg) error {
	if err := binary.Read(readBuffer, binary.BigEndian, &msg.MessageID); err != nil {
		h.logger.ErrorF("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.Reserved); err != nil {
		h.logger.ErrorF("Error reading message type:", err)
		return err
	}

	h.logger.InfoF("Received message %v", *msg)

	return nil
}
