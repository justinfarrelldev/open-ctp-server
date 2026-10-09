package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/justinfarrelldev/open-ctp-server/docs"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	gameplayevents "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/events"
	serverruntime "github.com/justinfarrelldev/open-ctp-server/internal/runtime"
	_ "github.com/lib/pq"
)

//	@title			Open Call to Power Server
//	@description	This is the open-source Call to Power and Call to Power 2 server project. This project is not sponsored, maintained or affiliated with Activision.

//	@contact.name	API Support
//	@contact.email	justinfarrellwebdev@gmail.com

type Server struct{}

var (
	port  = 9000
	sleep = flag.Duration("sleep", time.Second*5, "duration between changes in health")

	system = ""
)

//go:embed docs/swagger.json
var spec []byte

type gameplayRuntimeClock struct{}

func (gameplayRuntimeClock) Now() time.Time { return time.Now().UTC() }

type gameplayRuntimeIDs struct{ next atomic.Uint64 }

func (s *gameplayRuntimeIDs) Next() string {
	return fmt.Sprintf("gameplay-%d", s.next.Add(1))
}

type gameplayRuntimeRNG struct{}

func (gameplayRuntimeRNG) Next(_ string, state uint64) (uint64, uint64) {
	next := state + 0x9e3779b97f4a7c15
	value := next
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return next, value ^ (value >> 31)
}

func main() {
	if os.Getenv("SUPABASE_DB_URL") == "" {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Error loading .env file")
		}
	}

	db, err := sqlx.Open("postgres", os.Getenv("SUPABASE_DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("opened connection to database successfully")

	runtime := serverruntime.New(db, serverruntime.Config{
		Sources: engine.Sources{Clock: gameplayRuntimeClock{}, IDs: &gameplayRuntimeIDs{}, RNG: gameplayRuntimeRNG{}},
		Swagger: spec,
	})
	go dispatchGameplayEvents(runtime.Dispatcher)
	go runGameplayWorker("pending takeover", runtime.PendingTakeovers.RunOnce)
	go runGameplayWorker("ai driver", runtime.AIWorker.RunOnce)

	fmt.Printf("\nNow serving on port %d\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), runtime.Handler)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func dispatchGameplayEvents(dispatcher *gameplayevents.Dispatcher) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := dispatcher.DispatchOnce(context.Background()); err != nil {
			log.Printf("gameplay event outbox dispatch failed: %v", err)
		}
		<-ticker.C
	}
}

func runGameplayWorker(name string, run func(context.Context) error) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := run(context.Background()); err != nil {
			log.Printf("gameplay %s failed: %v", name, err)
		}
		<-ticker.C
	}
}
