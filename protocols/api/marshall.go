package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// unmarshallAnnounce parses the JSON data manually, and populates the AnnounceMsg struct
func unmarshallAnnounce(readBuffer *bytes.Reader, msg *AnnounceMsg) error {
	if err := binary.Read(readBuffer, binary.BigEndian, &msg.TTL); err != nil {
		fmt.Println("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.Reserved); err != nil {
		fmt.Println("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.DataType); err != nil {
		fmt.Println("Error reading message type:", err)
		return err
	}

	msgBuf := make([]byte, readBuffer.Len())

	if err := binary.Read(readBuffer, binary.BigEndian, &msgBuf); err != nil {
		fmt.Println("Error reading message type:", err)
		return err
	} else {
		msg.Data = string(msgBuf)
	}

	fmt.Printf("Received message %v\n", *msg)

	return nil
}

// unmarshallNotify parses the JSON data manually, and populates the NotifyMsg struct
func unmarshallNotify(readBuffer *bytes.Reader, msg *NotifyMsg) error {
	if err := binary.Read(readBuffer, binary.BigEndian, &msg.Reserved); err != nil {
		fmt.Println("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.DataType); err != nil {
		fmt.Println("Error reading message type:", err)
		return err
	}

	fmt.Printf("Received message %v\n", *msg)

	return nil
}

// marshallNotification parses the JSON data manually, and populates the NotificationMsg struct
func unmarshallNotification(readBuffer *bytes.Reader, msg *NotificationMsg) error {
	if err := binary.Read(readBuffer, binary.BigEndian, &msg.MessageID); err != nil {
		fmt.Println("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.DataType); err != nil {
		fmt.Println("Error reading message type:", err)
		return err
	}

	msgBuf := make([]byte, readBuffer.Len())

	if err := binary.Read(readBuffer, binary.BigEndian, &msgBuf); err != nil {
		fmt.Println("Error reading message type:", err)
		return err
	} else {
		msg.Data = string(msgBuf)
	}

	fmt.Printf("Received message %v\n", *msg)

	return nil
}

// unmarshallValidation parses the JSON data manually, and populates the ValidationMsg struct
func unmarshallValidation(readBuffer *bytes.Reader, msg *ValidationMsg) error {
	if err := binary.Read(readBuffer, binary.BigEndian, &msg.MessageID); err != nil {
		fmt.Println("Error reading message type:", err)
		return err
	}

	if err := binary.Read(readBuffer, binary.BigEndian, &msg.Reserved); err != nil {
		fmt.Println("Error reading message type:", err)
		return err
	}

	fmt.Printf("Received message %v\n", *msg)

	return nil
}
