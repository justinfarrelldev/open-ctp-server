//go:generate ./scripts/deps.sh

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	unitService "github.com/justinfarrelldev/open-ctp-server/units"
	"google.golang.org/grpc"

	"google.golang.org/grpc/health"
	"google.golang.org/grpc/reflection"

	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
)

type Server struct {
}

var (
	port  = 9000
	sleep = flag.Duration("sleep", time.Second*5, "duration between changes in health")

	system = "" // empty string represents the health of the system
)

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d", port)
	}

	fmt.Printf("Listening to TCP on port %d\n", port)

	s := grpc.NewServer()

	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(s, healthcheck)
	unitService.RegisterUnitsServer(s, &unitService.Server{})

	reflection.Register(s)

	go func() {
		// asynchronously inspect dependencies and toggle serving status as needed
		next := healthgrpc.HealthCheckResponse_SERVING

		for {
			healthcheck.SetServingStatus(system, next)

			if next == healthgrpc.HealthCheckResponse_SERVING {
				next = healthgrpc.HealthCheckResponse_NOT_SERVING
			} else {
				next = healthgrpc.HealthCheckResponse_SERVING
			}

			time.Sleep(*sleep)
		}
	}()

	if err := s.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC over port %d: %v", port, err)
	}
}
