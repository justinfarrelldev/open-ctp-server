package main

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"net/http"

	game "github.com/justinfarrelldev/open-ctp-server/internal/game"
	health "github.com/justinfarrelldev/open-ctp-server/internal/health"

	_ "github.com/justinfarrelldev/open-ctp-server/docs"

	"github.com/flowchartsman/swaggerui"

	"github.com/didip/tollbooth/v7"

	_ "github.com/lib/pq"
)

//	@title			Open Call to Power Server
//	@description	This is the open-source Call to Power and Call to Power 2 server project. This project is not sponsored, maintained or affiliated with Activision.

//	@contact.name	API Support
//	@contact.email	justinfarrellwebdev@gmail.com

type Server struct {
}

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

var (
	port  = 9000
	sleep = flag.Duration("sleep", time.Second*5, "duration between changes in health")

	system = "" // empty string represents the health of the system
)

//go:embed docs/swagger.json
var spec []byte

func main() {
	// Tollbooth
	message := Message{
		Status: "Request Failed",
		Body:   "The API is at capacity, try again later.",
	}
	jsonMessage, _ := json.Marshal(message)

	tollboothLimiter := tollbooth.NewLimiter(5, nil)
	tollboothLimiter.SetMessageContentType("application/json")
	tollboothLimiter.SetMessage(string(jsonMessage))

	// Postgres
	db, err := sql.Open("postgres", os.Getenv("SUPABASE_DB_URL"))
	if err != nil {
		log.Fatal(err)
	}

	// Handlers
	mux := http.NewServeMux()

	mux.Handle("/game/create_game", tollbooth.LimitFuncHandler(tollboothLimiter, func(w http.ResponseWriter, r *http.Request) {
		game.GameHandler(w, r, db)
	}))
	mux.Handle("/health", tollbooth.LimitFuncHandler(tollboothLimiter, health.HealthCheckHandler))
	mux.Handle("/docs/", http.StripPrefix("/docs", swaggerui.Handler(spec)))

	fmt.Printf("\nNow serving on port %d\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), mux)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
