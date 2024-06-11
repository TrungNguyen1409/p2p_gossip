package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/protocols/api"
	"net"
	"os"
)

func main() {
	var (
		announce    bool
		notify      bool
		destination string
		port        int
		message     string
	)

	flag.BoolVar(&announce, "a", false, "Send a GOSSIP_ANNOUNCE message")
	flag.BoolVar(&notify, "n", false, "Send a GOSSIP_NOTIFY message")
	flag.StringVar(&destination, "d", "", "GOSSIP host module IP")
	flag.IntVar(&port, "p", 0, "GOSSIP host module port")
	flag.StringVar(&message, "m", "", "GOSSIP host module port")

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
		fmt.Println("Destination must be specified.")
		os.Exit(1)
	}

	if port == 0 {
		fmt.Println("Port must be specified.")
		os.Exit(1)
	}

	if announce == true && notify == true {
		fmt.Println("Only announce or notify.")
		os.Exit(1)
	}

	if announce && message == "" {
		fmt.Println("Empty message for announce.")
		os.Exit(1)
	}

	address := fmt.Sprintf("%s:%d", destination, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("Failed to connect to %s: %v\n", address, err)
		os.Exit(1)
	}

	if announce {
		sendMessage(conn, api.GossipAnnounce, createAnnounceMessage(message))
	} else if notify {
		sendMessage(conn, api.GossipNotify, createNotifyMessage())
	}
}

func sendMessage(conn net.Conn, messageType uint16, message []byte) {
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
		fmt.Printf("Failed to send message: %v\n", err)
		return
	} else {
		fmt.Printf("Sent %d bytes for message type %d.\n", writeBuffer.Len(), messageType)
	}

	var response []byte
	n, err := conn.Read(response)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Received response: %s\n", string(response[:n]))
}

func createAnnounceMessage(message string) []byte {
	const (
		TTL      = uint8(1)
		RESERVED = uint8(1)
		DATATYPE = uint16(1)
	)

	var buffer bytes.Buffer

	_ = binary.Write(&buffer, binary.BigEndian, TTL)
	_ = binary.Write(&buffer, binary.BigEndian, RESERVED)
	_ = binary.Write(&buffer, binary.BigEndian, DATATYPE)
	_ = binary.Write(&buffer, binary.BigEndian, []byte(message))

	return buffer.Bytes()
}

func createNotifyMessage() []byte {
	const (
		RESERVED = uint16(1)
		DATATYPE = uint16(2)
	)

	var buffer bytes.Buffer

	_ = binary.Write(&buffer, binary.BigEndian, RESERVED)
	_ = binary.Write(&buffer, binary.BigEndian, DATATYPE)

	return buffer.Bytes()
}
