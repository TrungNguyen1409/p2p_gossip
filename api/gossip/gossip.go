package gossip

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	UnimplementedGossipServiceServer
}

func (s Server) Announce(context.Context, *GossipAnnounce) (*emptypb.Empty, error) {
	fmt.Println("Received announce request")

	return &emptypb.Empty{}, nil
}

func (s Server) Notify(context.Context, *GossipNotify) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s Server) Notification(context.Context, *GossipNotify) (*GossipValidation, error) {
	return &GossipValidation{
		Size:             10,
		GossipValidation: 10,
		MessageId:        10,
		Reserved:         10,
		IsValid:          true,
	}, nil
}
