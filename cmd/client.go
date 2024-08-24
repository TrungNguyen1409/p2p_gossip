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
		sendMessage(conn, enum.GossipAnnounce, createAnnounceMessage(message, uint8(ttl), datatype))
	} else if notify {
		sendMessage(conn, enum.GossipNotify, createNotifyMessage(datatype))
	}
}

func sendMessage(conn net.Conn, messageType uint16, message []byte) {
	logger := logging.NewCustomLogger()

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

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
