package main

import (
	"fmt"
	gossip2 "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/api/gossip"
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

	gossipServer := &gossip2.Server{}

	grpcServer := grpc.NewServer()

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	gossip2.RegisterGossipServiceServer(grpcServer, gossipServer)

	fmt.Printf("Server started listenting on %s\n", apiAddress)

	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("Can not serve %v", err)
	}
}
