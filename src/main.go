//go:generate ../scripts/deps.sh

package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	port := 9000

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d", port)
	}

	fmt.Printf("Listening to TCP on port %d\n", port)

	grpcServer := grpc.NewServer()

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC over port %d: %v", port, err)
	}
}
