package api

import (
	"fmt"
	"net"
)

func sendResponse(conn net.Conn, s string) {
	response := []byte(s)
	write, err := conn.Write(response)
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
	fmt.Printf("Sent %d bytes. Message: %s", write, s)
}
