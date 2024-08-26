package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"net/http"
)

type Server struct {
}

var (
	port  = 9000
	sleep = flag.Duration("sleep", time.Second*5, "duration between changes in health")

	system = "" // empty string represents the health of the system
)

func main() {

	mux := http.NewServeMux()

	// mux.HandleFunc("/units", units.GetAllUnitInfo)
	// mux.HandleFunc("/units/", units.GetUnitInfo)

	fmt.Printf("\nNow serving on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
