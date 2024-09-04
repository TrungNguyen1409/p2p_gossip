package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
)

func main() {

	logger := logging.NewCustomLogger()

	var (
		announce    bool
		notify      bool
		destination string
		port        int
		message     string
		ttl         int
		datatype    int
	)

	flag.BoolVar(&announce, "a", false, "Send a GOSSIP_ANNOUNCE message")
	flag.BoolVar(&notify, "n", false, "Send a GOSSIP_NOTIFY message")
	flag.StringVar(&destination, "d", "", "GOSSIP host module IP")
	flag.IntVar(&port, "p", 0, "GOSSIP host module port")
	flag.IntVar(&ttl, "ttl", 1, "GOSSIP host module port")
	flag.StringVar(&message, "m", "", "GOSSIP_ANNOUNCE payload message")
	flag.IntVar(&datatype, "t", 1, "GOSSIP_NOTIFY datatype")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		_, _ = fmt.Fprintf(os.Stderr, `
Examples:
  Send a GOSSIP_ANNOUNCE message:
    %s -a -d 127.0.0.1 -p 9001 announce_message

  Send a GOSSIP_NOTIFY message:
    %s -n -d 127.0.0.1 -p 9001 -m notify_message
`, os.Args[0], os.Args[0])
	}

	flag.Parse()

	if destination == "" {
		logger.Info("Destination must be specified.")
		os.Exit(1)
	}

	if port == 0 {
		logger.Info("Port must be specified.")
		os.Exit(1)
	}

	if announce == true && notify == true {
		logger.Info("Only announce or notify.")
		os.Exit(1)
	}

	if announce && message == "" {
		logger.Info("Empty message for announce.")
		os.Exit(1)
	}

	address := fmt.Sprintf("%s:%d", destination, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.ErrorF("Failed to connect to %s: %v\n", address, err)
		os.Exit(1)
	}

	if announce {
		sendMessage(conn, enum.GossipAnnounce, createAnnounceMessage(message, uint8(ttl), datatype), logger)
	} else if notify {
		handleConnection(conn, datatype, logger)
	}
}

func handleConnection(conn net.Conn, datatype int, logger *logging.Logger) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.ErrorF("Failed to close connection: %v\n", err)
		}
	}(conn)

	logger.InfoF("Accepted connection from %s\n", conn.RemoteAddr())

	message := createNotifyMessage(datatype)
	sendMessageAndWaitForResponse(conn, enum.GossipNotify, message, logger)
}

func sendMessageAndWaitForResponse(conn net.Conn, messageType uint16, message []byte, logger *logging.Logger) {
	var writeBuffer bytes.Buffer

	_ = binary.Write(&writeBuffer, binary.BigEndian, uint16(len(message)+4))
	_ = binary.Write(&writeBuffer, binary.BigEndian, messageType)
	_ = binary.Write(&writeBuffer, binary.BigEndian, message)
	_, err := conn.Write(writeBuffer.Bytes())
	if err != nil {
		logger.InfoF("Failed to send message: %v\n", err)
		return
	} else {
		logger.InfoF("Sent %d bytes for message type %d.\n", writeBuffer.Len(), messageType)
	}

	for {
		messageBuffer := make([]byte, 1024)
		_, err := conn.Read(messageBuffer)

		logger.DebugF("Reading from %s\n", conn.RemoteAddr())

		if err != nil {
			if err.Error() == "EOF" {
				logger.Info("Connection closed by peer.")
				return
			}
			logger.ErrorF("Error reading response: %v\n", err)
			return
		}

		reader := bytes.NewReader(messageBuffer)

		var size uint16
		var messageType uint16

		var msg enum.NotificationMsg

		// Read the size
		if err = binary.Read(reader, binary.BigEndian, &size); err != nil {
			logger.ErrorF("Error reading size: %v\n", err)
		}

		// Read the message type
		if err = binary.Read(reader, binary.BigEndian, &messageType); err != nil {
			logger.ErrorF("Error reading message type: %v\n", err)
		}

		// marshallNotification parses the JSON data manually, and populates the NotificationMsg struct
		if err := binary.Read(reader, binary.BigEndian, &msg.MessageID); err != nil {
			logger.ErrorF("Error reading message type:", err)
		}

		if err := binary.Read(reader, binary.BigEndian, &msg.DataType); err != nil {
			logger.ErrorF("Error reading message type:", err)
		}

		msgBuf := make([]byte, reader.Len())

		if err := binary.Read(reader, binary.BigEndian, &msgBuf); err != nil {
			logger.ErrorF("Error reading message type:", err)
		} else {
			msg.Data = string(msgBuf)
		}

		logger.InfoF("Received message %+v", msg)
	}
}

func sendMessage(conn net.Conn, messageType uint16, message []byte, logger *logging.Logger) {
	var writeBuffer bytes.Buffer

	_ = binary.Write(&writeBuffer, binary.BigEndian, uint16(len(message)+4))
	_ = binary.Write(&writeBuffer, binary.BigEndian, messageType)
	_ = binary.Write(&writeBuffer, binary.BigEndian, message)
	_, err := conn.Write(writeBuffer.Bytes())
	if err != nil {
		logger.InfoF("Failed to send message: %v\n", err)
		return
	} else {
		logger.InfoF("Sent %d bytes for message type %d.\n", writeBuffer.Len(), messageType)
	}
}

func createAnnounceMessage(message string, ttl uint8, datatype int) []byte {

	var (
		TTL      = ttl
		RESERVED = uint8(1)
		DATATYPE = enum.Datatype(datatype)
	)

	var buffer bytes.Buffer

	_ = binary.Write(&buffer, binary.BigEndian, TTL)
	_ = binary.Write(&buffer, binary.BigEndian, RESERVED)
	_ = binary.Write(&buffer, binary.BigEndian, DATATYPE)
	_ = binary.Write(&buffer, binary.BigEndian, []byte(message))

	return buffer.Bytes()
}

func createNotifyMessage(datatype int) []byte {
	var (
		RESERVED = uint16(1)
		DATATYPE = enum.Datatype(datatype)
	)

	var buffer bytes.Buffer

	_ = binary.Write(&buffer, binary.BigEndian, RESERVED)
	_ = binary.Write(&buffer, binary.BigEndian, DATATYPE)

	return buffer.Bytes()
}

func unmarshallMessage(conn net.Conn, logger *logging.Logger) error {
	messageBuffer := make([]byte, 1024)
	_, err := conn.Read(messageBuffer)

	logger.DebugF("Reading from %s\n", conn.RemoteAddr())

	if err != nil {
		return fmt.Errorf("Error reading from connection: %v\n", err)
	}

	reader := bytes.NewReader(messageBuffer)

	var size uint16
	var messageType uint16

	var msg enum.NotificationMsg

	// Read the size
	if err = binary.Read(reader, binary.BigEndian, &size); err != nil {
		logger.ErrorF("Error reading size: %v\n", err)
	}

	// Read the message type
	if err = binary.Read(reader, binary.BigEndian, &messageType); err != nil {
		return fmt.Errorf("Error reading message type: %v\n", err)
	}

	// marshallNotification parses the JSON data manually, and populates the NotificationMsg struct
	if err := binary.Read(reader, binary.BigEndian, &msg.MessageID); err != nil {
		return fmt.Errorf("Error reading message type:", err)
	}

	if err := binary.Read(reader, binary.BigEndian, &msg.DataType); err != nil {
		return fmt.Errorf("Error reading message type:", err)
	}

	msgBuf := make([]byte, reader.Len())

	if err := binary.Read(reader, binary.BigEndian, &msgBuf); err != nil {
		return fmt.Errorf("Error reading message type:", err)
	} else {
		msg.Data = string(msgBuf)
	}

	logger.InfoF("Received message %v", msg)

	return nil
}
