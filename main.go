package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"net/http"

	game "github.com/justinfarrelldev/open-ctp-server/internal/game"
	health "github.com/justinfarrelldev/open-ctp-server/internal/health"

	_ "github.com/justinfarrelldev/open-ctp-server/docs"

	"github.com/flowchartsman/swaggerui"
)

//	@title			Open Call to Power Server
//	@description	This is the open-source Call to Power and Call to Power 2 server project. This project is not sponsored, maintainer or affiliated with Activision.

//	@contact.name	API Support
//	@contact.email	justinfarrellwebdev@gmail.com

type Server struct {
}

var (
	port  = 9000
	sleep = flag.Duration("sleep", time.Second*5, "duration between changes in health")

	system = "" // empty string represents the health of the system
)

//go:embed docs/swagger.json
var spec []byte

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/game/create_game", game.GameHandler)
	mux.HandleFunc("/health", health.HealthCheckHandler)
	mux.Handle("/docs/", http.StripPrefix("/docs", swaggerui.Handler(spec)))

	fmt.Printf("\nNow serving on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
