package api

import (
	"bytes"
	"encoding/binary"
)

// unmarshallAnnounce parses the JSON data manually, and populates the AnnounceMsg struct
func (handler *Handler) unmarshallAnnounce(readBuffer *bytes.Reader, msg *AnnounceMsg) error {
	if err := binary.Read(readBuffer, binary.BigEndian, &msg.TTL); err != nil {
		handler.logger.ErrorF("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.Reserved); err != nil {
		handler.logger.ErrorF("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.DataType); err != nil {
		handler.logger.ErrorF("Error reading message type:", err)
		return err
	}

	msgBuf := make([]byte, readBuffer.Len())

	if err := binary.Read(readBuffer, binary.BigEndian, &msgBuf); err != nil {
		handler.logger.ErrorF("Error reading message type:", err)
		return err
	} else {
		msg.Data = string(msgBuf)
	}

	handler.logger.InfoF("Received message %v", *msg)

	return nil
}

// unmarshallNotify parses the JSON data manually, and populates the NotifyMsg struct
func (handler *Handler) unmarshallNotify(readBuffer *bytes.Reader, msg *NotifyMsg) error {
	if err := binary.Read(readBuffer, binary.BigEndian, &msg.Reserved); err != nil {
		handler.logger.ErrorF("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.DataType); err != nil {
		handler.logger.ErrorF("Error reading message type:", err)
		return err
	}

	handler.logger.InfoF("Received message %v", *msg)

	return nil
}

// marshallNotification parses the JSON data manually, and populates the NotificationMsg struct
func (handler *Handler) unmarshallNotification(readBuffer *bytes.Reader, msg *NotificationMsg) error {
	if err := binary.Read(readBuffer, binary.BigEndian, &msg.MessageID); err != nil {
		handler.logger.ErrorF("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.DataType); err != nil {
		handler.logger.ErrorF("Error reading message type:", err)
		return err
	}

	msgBuf := make([]byte, readBuffer.Len())

	if err := binary.Read(readBuffer, binary.BigEndian, &msgBuf); err != nil {
		handler.logger.ErrorF("Error reading message type:", err)
		return err
	} else {
		msg.Data = string(msgBuf)
	}

	handler.logger.InfoF("Received message %v", *msg)

	return nil
}

// unmarshallValidation parses the JSON data manually, and populates the ValidationMsg struct
func (handler *Handler) unmarshallValidation(readBuffer *bytes.Reader, msg *ValidationMsg) error {
	if err := binary.Read(readBuffer, binary.BigEndian, &msg.MessageID); err != nil {
		handler.logger.ErrorF("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.Reserved); err != nil {
		handler.logger.ErrorF("Error reading message type:", err)
		return err
	}

	handler.logger.InfoF("Received message %v", *msg)

	return nil
}
