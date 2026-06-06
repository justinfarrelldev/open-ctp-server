package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"net/http"

	"github.com/didip/tollbooth/v7"
	"github.com/flowchartsman/swaggerui"
	"github.com/jmoiron/sqlx"
	account "github.com/justinfarrelldev/open-ctp-server/internal/account"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
	game "github.com/justinfarrelldev/open-ctp-server/internal/game"
	health "github.com/justinfarrelldev/open-ctp-server/internal/health"
	lobby "github.com/justinfarrelldev/open-ctp-server/internal/lobby"
	ratelimit "github.com/justinfarrelldev/open-ctp-server/internal/ratelimit"

	"github.com/joho/godotenv"
	_ "github.com/justinfarrelldev/open-ctp-server/docs"
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

func main() {
	if os.Getenv("SUPABASE_DB_URL") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file")
		}
	}

	authWriteLimiter := ratelimit.NewAuthWriteLimiter()
	accountLimiter := ratelimit.NewAccountLimiter()
	lobbyGameLimiter := ratelimit.NewLobbyGameLimiter()
	healthLimiter := ratelimit.NewHealthLimiter()

	db, err := sqlx.Open("postgres", os.Getenv("SUPABASE_DB_URL"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("opened connection to database successfully")

	sessionStore := auth.NewSessionStore(db)
	mux := http.NewServeMux()

	mux.Handle("/game/create_game", tollbooth.LimitFuncHandler(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) {
		game.GameHandler(w, r, db)
	}))

	mux.Handle("/account/create_account", tollbooth.LimitFuncHandler(authWriteLimiter, func(w http.ResponseWriter, r *http.Request) {
		account.CreateAccountHandler(w, r, db, sessionStore)
	}))
	mux.Handle("/account/login", tollbooth.LimitFuncHandler(authWriteLimiter, func(w http.ResponseWriter, r *http.Request) {
		account.LoginAccountHandler(w, r, db, sessionStore)
	}))
	mux.Handle("/account/reset_password", tollbooth.LimitFuncHandler(authWriteLimiter, func(w http.ResponseWriter, r *http.Request) {
		account.ResetPasswordHandler(w, r, db, sessionStore)
	}))
	mux.Handle("/account/get_account", tollbooth.LimitFuncHandler(accountLimiter, func(w http.ResponseWriter, r *http.Request) {
		account.GetAccountHandler(w, r, db, sessionStore)
	}))
	mux.Handle("/account/update_account", tollbooth.LimitFuncHandler(accountLimiter, func(w http.ResponseWriter, r *http.Request) {
		account.UpdateAccountHandler(w, r, db, sessionStore)
	}))
	mux.Handle("/account/delete_account", tollbooth.LimitFuncHandler(accountLimiter, func(w http.ResponseWriter, r *http.Request) {
		account.DeleteAccountHandler(w, r, db, sessionStore)
	}))

	mux.Handle("/lobby/create_lobby", tollbooth.LimitFuncHandler(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) {
		lobby.CreateLobbyHandler(w, r, db, sessionStore)
	}))
	mux.Handle("/lobby/get_lobby", tollbooth.LimitFuncHandler(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) {
		lobby.GetLobbyHandler(w, r, db, sessionStore)
	}))
	mux.Handle("/lobby/update_lobby", tollbooth.LimitFuncHandler(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) {
		lobby.UpdateLobbyHandler(w, r, db, sessionStore)
	}))
	mux.Handle("/lobby/delete_lobby", tollbooth.LimitFuncHandler(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) {
		lobby.DeleteLobbyHandler(w, r, db, sessionStore)
	}))
	mux.Handle("/lobby/join_lobby", tollbooth.LimitFuncHandler(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) {
		lobby.JoinLobbyHandler(w, r, db, sessionStore)
	}))
	mux.Handle("/lobby/leave_lobby", tollbooth.LimitFuncHandler(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) {
		lobby.LeaveLobbyHandler(w, r, db, sessionStore)
	}))
	mux.Handle("/lobby/connect_lobby", tollbooth.LimitFuncHandler(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) {
		lobby.ConnectLobbyHandler(w, r, db, sessionStore)
	}))
	mux.Handle("/lobby/send_chat", tollbooth.LimitFuncHandler(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) {
		lobby.SendChatHandler(w, r, db, sessionStore)
	}))

	mux.Handle("/health", tollbooth.LimitFuncHandler(healthLimiter, health.HealthCheckHandler))
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
