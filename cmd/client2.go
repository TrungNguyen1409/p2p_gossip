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
		destination string
		port        int
		datatype    int
	)

	flag.StringVar(&destination, "d", "", "GOSSIP host module IP")
	flag.IntVar(&port, "p", 0, "GOSSIP host module port")
	flag.IntVar(&datatype, "t", 1, "GOSSIP_NOTIFY datatype")

	flag.Parse()

	if port == 0 {
		logger.Info("Port must be specified.")
		os.Exit(1)
	}

	// Start server
	address := fmt.Sprintf("%s:%d", destination, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.ErrorF("Failed to start server on port %d: %v\n", port, err)
		os.Exit(1)
	}

	logger.InfoF("Server started on port %d, waiting for connections...\n", port)

	handleConnection(conn, datatype, logger)
}

func handleConnection(conn net.Conn, datatype int, logger *logging.Logger) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.ErrorF("Failed to close connection: %v\n", err)
		}
	}(conn)

	logger.InfoF("Accepted connection from %s\n", conn.RemoteAddr())

	// Send Notify message immediately after connection is established
	message := createNotifyMessage2(datatype)
	sendMessage2(conn, enum.GossipNotify, message)

	// Wait for response after sending Notify message
	waitForResponse(conn, logger)
}

func waitForResponse(conn net.Conn, logger *logging.Logger) {
	var response = make([]byte, 1024) // Buffer for response

	for {
		n, err := conn.Read(response)
		if err != nil {
			if err.Error() == "EOF" {
				logger.Info("Connection closed by peer.")
				return
			}
			logger.ErrorF("Error reading response: %v\n", err)
			return
		}

		logger.InfoF("Received response: %s\n", string(response[:n]))
	}
}

func sendMessage2(conn net.Conn, messageType uint16, message []byte) {
	logger := logging.NewCustomLogger()

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

	var response []byte
	n, err := conn.Read(response)
	if err != nil {
		logger.InfoF("Error reading response: %v\n", err)
		return
	}

	logger.InfoF("Received response: %s\n", string(response[:n]))
}

func createNotifyMessage2(datatype int) []byte {
	var (
		RESERVED = uint16(1)
		DATATYPE = enum.Datatype(datatype)
	)

	var buffer bytes.Buffer

	_ = binary.Write(&buffer, binary.BigEndian, RESERVED)
	_ = binary.Write(&buffer, binary.BigEndian, DATATYPE)

	return buffer.Bytes()
}
