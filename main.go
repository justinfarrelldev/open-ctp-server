//go:generate ./scripts/deps.sh

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
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

func getAllUnitInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got request to /units!")
	fmt.Printf("\nHttp request: %v", r)
}

func getUnitInfo(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	unitType := parts[len(parts)-1]

	fmt.Printf("\nGot request to /units/%v!", unitType)
	fmt.Printf("\nHttp request: %v", r)
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/units", getAllUnitInfo)
	mux.HandleFunc("/units/", getUnitInfo)

	fmt.Printf("\nNow serving on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
