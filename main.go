//go:generate ../scripts/deps.sh

package main

import (
	"fmt"
	"log"
	"net"

	unitService "github.com/justinfarrelldev/open-ctp-server/units"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
}

func main() {
	port := 9000

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d", port)
	}

	fmt.Printf("Listening to TCP on port %d\n", port)

	grpcServer := grpc.NewServer()

	unitService.RegisterUnitsServer(grpcServer, &unitService.Server{})

	reflection.Register(grpcServer)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC over port %d: %v", port, err)
	}
}
