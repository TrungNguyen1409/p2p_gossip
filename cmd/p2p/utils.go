package p2p

import (
	"github.com/golang/protobuf/proto"
	pb "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/proto"
	"net"
)

// serialize converts a GossipMessage into a byte slice.
func serialize(msg *pb.GossipMessage) ([]byte, error) {
	return proto.Marshal(msg)
}

// deserialize converts a byte slice into a GossipMessage.
func deserialize(data []byte) (*pb.GossipMessage, error) {
	var msg pb.GossipMessage
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// send sends a GossipMessage to the specified address over TCP.
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

	data, err := serialize(msg)
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}
