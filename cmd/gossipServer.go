package main

import (
	"fmt"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/internal/gossip"
	"google.golang.org/grpc/reflection"
	"log"
	"net"

	"github.com/robfig/config"
	"google.golang.org/grpc"
)

func main() {
	c, err := config.ReadDefault("configs/config.ini")
	if err != nil {
		log.Fatalf("Can not find config.ini %v", err)
	}

	apiAddress, err := c.String("gossip", "api_address")
	if err != nil {
		log.Fatalf("Can not read from config.ini %v", err)
	}

	lis, err := net.Listen("tcp", apiAddress)
	if err != nil {
		log.Fatalf("Can not listen %v", err)
	}

	gossipServer := &gossip.Server{}

	grpcServer := grpc.NewServer()

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	gossip.RegisterGossipServiceServer(grpcServer, gossipServer)

	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("Can not serve %v", err)
	}

	fmt.Printf("Server started listenting on %s\n", apiAddress)
}
