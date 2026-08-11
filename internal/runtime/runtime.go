package runtime

import (
	"net/http"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/flowchartsman/swaggerui"
	"github.com/jmoiron/sqlx"
	account "github.com/justinfarrelldev/open-ctp-server/internal/account"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
	game "github.com/justinfarrelldev/open-ctp-server/internal/game"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/ai"
	gameplaycombat "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/combat"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	gameplayevents "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/events"
	gameplayhttp "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/http"
	gameplayservice "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/service"
	gameplaystore "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
	health "github.com/justinfarrelldev/open-ctp-server/internal/health"
	lobby "github.com/justinfarrelldev/open-ctp-server/internal/lobby"
	ratelimit "github.com/justinfarrelldev/open-ctp-server/internal/ratelimit"
)

type Config struct {
	Sources          engine.Sources
	StartSeed        func() (int64, error)
	Swagger          []byte
	NewLimiter       func(string) *limiter.Limiter
	ObserveAICommand func(ai.Submission)
}

type Runtime struct {
	Handler    http.Handler
	Dispatcher *gameplayevents.Dispatcher
	Deadlines  *gameplayservice.DeadlineWorker
	AIDriver   *ai.Driver
}

func New(db *sqlx.DB, config Config) *Runtime {
	sessions := auth.NewSessionStore(db)
	broker := gameplayevents.NewBroker()
	service := gameplayservice.New(gameplaystore.New(db), config.Sources)
	gateway := gameplayservice.NewAIGateway(service)
	dispatcher := gameplayevents.NewDispatcher(db, broker)
	newLimiter := func(name string, standard func() *limiter.Limiter) *limiter.Limiter {
		if config.NewLimiter != nil {
			return config.NewLimiter(name)
		}
		return standard()
	}
	authWriteLimiter := newLimiter("auth-write", ratelimit.NewAuthWriteLimiter)
	accountLimiter := newLimiter("account", ratelimit.NewAccountLimiter)
	lobbyGameLimiter := newLimiter("lobby-game", ratelimit.NewLobbyGameLimiter)
	gameplayCommandLimiter := newLimiter("gameplay-command", ratelimit.NewGameplayCommandLimiter)
	lobbySSELimiter := newLimiter("lobby-sse", ratelimit.NewLobbySSELimiter)
	healthLimiter := newLimiter("health", ratelimit.NewHealthLimiter)
	limit := func(limiter *limiter.Limiter, handler http.HandlerFunc) http.Handler {
		return tollbooth.LimitFuncHandler(limiter, handler)
	}
	mux := http.NewServeMux()

	mux.Handle("/game/create_game", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.CreateGameHandler(w, r, db, sessions) }))
	mux.Handle("/game/get_game", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.GetGameHandler(w, r, db, sessions) }))
	mux.Handle("/game/update_game", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.UpdateGameHandler(w, r, db, sessions) }))
	mux.Handle("/game/delete_game", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.DeleteGameHandler(w, r, db, sessions) }))
	mux.Handle("/game/rename_game", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.RenameGameHandler(w, r, db, sessions) }))
	mux.Handle("/game/lock_game", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.LockGameHandler(w, r, db, sessions) }))
	mux.Handle("/game/set_public", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.SetGamePublicHandler(w, r, db, sessions) }))
	mux.Handle("/game/update_style", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.UpdateGameStyleHandler(w, r, db, sessions) }))
	mux.Handle("/game/update_rules", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.UpdateGameRulesHandler(w, r, db, sessions) }))
	mux.Handle("/game/update_exclusions", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.UpdateGameExclusionsHandler(w, r, db, sessions) }))
	mux.Handle("/game/add_ai", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.AddAIHandler(w, r, db, sessions) }))
	mux.Handle("/game/kick_player", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.KickGamePlayerHandler(w, r, db, sessions) }))
	mux.Handle("/game/select_country", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.SelectGameCountryHandler(w, r, db, sessions) }))
	mux.Handle("/game/update_starting_settlers", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) {
		game.UpdateGameStartingSettlersHandler(w, r, db, sessions)
	}))
	mux.Handle("/game/update_starting_gold", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.UpdateGameStartingGoldHandler(w, r, db, sessions) }))
	mux.Handle("/game/update_ping", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.UpdateGamePingHandler(w, r, db, sessions) }))
	if config.StartSeed == nil {
		mux.Handle("/game/start_game", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.StartGameHandler(w, r, db, sessions) }))
	} else {
		mux.Handle("/game/start_game", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) {
			if err := game.StartGameWithSeedAndSources(w, r, db, sessions, config.StartSeed, config.Sources); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}))
	}
	commandHandler := gameplayhttp.NewHandler(service, sessions)
	combatHandler := gameplaycombat.NewHandler(service, sessions)
	eventHandler := gameplayevents.NewHandler(db, sessions, visibility.NewProjector(), broker)
	mux.Handle("POST /game/{game_id}/combat", limit(gameplayCommandLimiter, combatHandler.ServeHTTP))
	mux.Handle("POST /game/{game_id}/{command}", limit(gameplayCommandLimiter, commandHandler.ServeHTTP))
	mux.Handle("GET /game/{game_id}/events", limit(lobbySSELimiter, eventHandler.ServeHTTP))
	mux.Handle("GET /world/{game_id}", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.GetWorldHandler(w, r, db, sessions) }))
	mux.Handle("GET /world/{game_id}/space", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.GetWorldSpaceLayerHandler(w, r, db, sessions) }))
	mux.Handle("GET /world/{game_id}/ratings", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.GetWorldRatingsHandler(w, r, db, sessions) }))
	mux.Handle("GET /world/{game_id}/pollution", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { game.GetWorldPollutionHandler(w, r, db, sessions) }))
	mux.Handle("/account/create_account", limit(authWriteLimiter, func(w http.ResponseWriter, r *http.Request) { account.CreateAccountHandler(w, r, db, sessions) }))
	mux.Handle("/account/login", limit(authWriteLimiter, func(w http.ResponseWriter, r *http.Request) { account.LoginAccountHandler(w, r, db, sessions) }))
	mux.Handle("/account/reset_password", limit(authWriteLimiter, func(w http.ResponseWriter, r *http.Request) { account.ResetPasswordHandler(w, r, db, sessions) }))
	mux.Handle("/account/get_account", limit(accountLimiter, func(w http.ResponseWriter, r *http.Request) { account.GetAccountHandler(w, r, db, sessions) }))
	mux.Handle("/account/update_account", limit(accountLimiter, func(w http.ResponseWriter, r *http.Request) { account.UpdateAccountHandler(w, r, db, sessions) }))
	mux.Handle("/account/delete_account", limit(accountLimiter, func(w http.ResponseWriter, r *http.Request) { account.DeleteAccountHandler(w, r, db, sessions) }))
	mux.Handle("/lobby/create_lobby", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { lobby.CreateLobbyHandler(w, r, db, sessions) }))
	mux.Handle("/lobby/get_lobby", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { lobby.GetLobbyHandler(w, r, db, sessions) }))
	mux.Handle("/lobby/update_lobby", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { lobby.UpdateLobbyHandler(w, r, db, sessions) }))
	mux.Handle("/lobby/delete_lobby", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { lobby.DeleteLobbyHandler(w, r, db, sessions) }))
	mux.Handle("/lobby/join_lobby", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { lobby.JoinLobbyHandler(w, r, db, sessions) }))
	mux.Handle("/lobby/leave_lobby", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { lobby.LeaveLobbyHandler(w, r, db, sessions) }))
	mux.Handle("/lobby/connect_lobby", limit(lobbySSELimiter, func(w http.ResponseWriter, r *http.Request) { lobby.ConnectLobbyHandler(w, r, db, sessions) }))
	mux.Handle("/lobby/send_chat", limit(lobbyGameLimiter, func(w http.ResponseWriter, r *http.Request) { lobby.SendChatHandler(w, r, db, sessions) }))
	mux.Handle("/health", limit(healthLimiter, health.HealthCheckHandler))
	mux.Handle("/docs/", http.StripPrefix("/docs", swaggerui.Handler(config.Swagger)))
	deadlines := gameplayservice.NewDeadlineWorker(service)
	deadlines.AfterCommit = dispatcher.DispatchOnce
	return &Runtime{Handler: mux, Dispatcher: dispatcher, Deadlines: deadlines, AIDriver: ai.NewDriver(gateway, gateway, ai.Config{ActionBudget: 1, History: gateway, Observe: config.ObserveAICommand})}
}
