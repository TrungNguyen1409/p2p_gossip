package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
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

}
