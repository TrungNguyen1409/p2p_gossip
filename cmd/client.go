package main

import (
	"context"
	"flag"
	"fmt"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/api/gossip"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"

	"google.golang.org/grpc"
)

func main() {
	var (
		announce    bool
		notify      bool
		destination string
		port        int
	)

	flag.BoolVar(&announce, "a", false, "Send a GOSSIP_ANNOUNCE message")
	flag.BoolVar(&notify, "n", false, "Send a GOSSIP_NOTIFY and subsequent VALIDATION")
	flag.StringVar(&destination, "d", "", "GOSSIP host module IP")
	flag.IntVar(&port, "p", 9001, "GOSSIP host module port")
	flag.Parse()

	address := fmt.Sprintf("%s:%d", destination, port)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v\n", err)
	}

	defer func(conn *grpc.ClientConn) {
		err = conn.Close()
		if err != nil {
			log.Fatalf("could not close connection: %v\n", err)
		}

	}(conn)

	client := gossip.NewGossipServiceClient(conn)

	if announce {
		sendAnnounce(client)
	}

	if notify {
		// Implement notify handling
	}
}

func sendAnnounce(client gossip.GossipServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	message := &gossip.GossipAnnounce{
		Size:           1,
		GossipAnnounce: 2,
		Ttl:            3,
		Reserved:       4,
		DataType:       5,
		Data:           []byte("data"),
	}

	_, err := client.Announce(ctx, message)

	if err != nil {
		log.Fatalf("Announce failed: %v", err)
	}
}
