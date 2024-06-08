package p2p

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"sync"
)

func (msg *GossipMessage) Serialize() ([]byte, error) {
	return proto.Marshal(msg)
}

func Deserialize(data []byte) (*GossipMessage, error) {
	var msg GossipMessage
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// SendMessage sends a message to a peer
func SendMessage(address string, msg *GossipMessage) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	data, err := msg.Serialize()
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}

// p2pAddress instead of + port
// ListenForMessages listens for incoming messages from peers
func ListenForMessages(p2pAddress string, msgHandler func(*GossipMessage), wg *sync.WaitGroup) {
	ln, err := net.Listen("tcp", p2pAddress)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}

	fmt.Println("P2P Server is listening on", ln.Addr())

	defer func(ln net.Listener) {
		err := ln.Close()
		if err != nil {
			log.Fatalf("Error closing TCP server: %v", err)
		}
	}(ln)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}

		go handleConnection(conn, msgHandler)
	}
}

func handleConnection(conn net.Conn, msgHandler func(*GossipMessage)) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Failed to read from connection:", err)
		return
	}

	msg, err := Deserialize(buf[:n])
	if err != nil {
		log.Println("Failed to deserialize message:", err)
		return
	}

	msgHandler(msg)
}
