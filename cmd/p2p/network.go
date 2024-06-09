package p2p

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"sync"

	pb "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/proto"
)

func Serialize(msg *pb.GossipMessage) ([]byte, error) {
	return proto.Marshal(msg)
}

func Deserialize(data []byte) (*pb.GossipMessage, error) {
	var msg pb.GossipMessage
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// SendMessage sends a message to a peer
func send(address string, msg *pb.GossipMessage) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	data, err := Serialize(msg)
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}

// ListenForMessages listens for incoming messages from peers
func listen(p2pAddress string, msgHandler func(*pb.GossipMessage), wg *sync.WaitGroup) {
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
		// Increment the WaitGroup counter
		wg.Add(1)
		go HandleConnection(conn, msgHandler)
	}
}

func HandleConnection(conn net.Conn, msgHandler func(*pb.GossipMessage)) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v\n", err)
			fmt.Printf("Connection closed!\n----------------------------------\n")
		}
	}(conn)

	fmt.Printf("\n----------------------------------\nOpen connection with %s\n", conn.RemoteAddr())

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
