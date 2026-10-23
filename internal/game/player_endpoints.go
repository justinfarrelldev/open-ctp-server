package game

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	world "github.com/justinfarrelldev/open-ctp-server/internal/world"
)

const ErrorNegativePing = "ping must be greater than or equal to 0"

type KickGamePlayerArgs struct {
	GameId    int64  `json:"game_id"`
	SessionId *int64 `json:"session_id"`
	AccountId int64  `json:"account_id"`
}

func (a KickGamePlayerArgs) GetSessionID() *int64 {
	return a.SessionId
}

type SelectGameCountryArgs struct {
	GameId    int64  `json:"game_id"`
	SessionId *int64 `json:"session_id"`
	Country   string `json:"country"`
}

func (a SelectGameCountryArgs) GetSessionID() *int64 {
	return a.SessionId
}

type UpdateGameStartingSettlersArgs struct {
	GameId           int64  `json:"game_id"`
	SessionId        *int64 `json:"session_id"`
	StartingSettlers int    `json:"starting_settlers"`
}

func (a UpdateGameStartingSettlersArgs) GetSessionID() *int64 {
	return a.SessionId
}

type UpdateGameStartingGoldArgs struct {
	GameId       int64  `json:"game_id"`
	SessionId    *int64 `json:"session_id"`
	StartingGold int    `json:"starting_gold"`
}

func (a UpdateGameStartingGoldArgs) GetSessionID() *int64 {
	return a.SessionId
}

type UpdateGamePingArgs struct {
	GameId    int64  `json:"game_id"`
	SessionId *int64 `json:"session_id"`
	Ping      int    `json:"ping"`
}

func (a UpdateGamePingArgs) GetSessionID() *int64 {
	return a.SessionId
}

type StartGameArgs struct {
	GameId    int64  `json:"game_id"`
	SessionId *int64 `json:"session_id"`
}

func (a StartGameArgs) GetSessionID() *int64 {
	return a.SessionId
}

func updateStartingValue(w http.ResponseWriter, db *sqlx.DB, store *auth.SessionStore, sessionBody auth.SessionIDBody, gameID int64, value int, errorMessage string, column string) error {
	session, err := requireGameSession(w, store, sessionBody)
	if err != nil {
		return err
	}
	if value < 0 {
		return writeGameStatusError(w, http.StatusBadRequest, errorMessage)
	}

	tx, err := beginGameTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting game update: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	game, err := getGameByIDForUpdate(tx, gameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("no game exists with the ID %d", gameID))
		}
		return err
	}
	if err := ensureGameOwner(w, session, game); err != nil {
		return err
	}
	if err := ensureGameSetupMutable(w, game); err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE games SET %s = $1, updated_at = NOW() WHERE id = $2", column)
	if _, err := tx.Exec(query, value, gameID); err != nil {
		return fmt.Errorf("an error occurred while updating the game with the ID %d: %v", gameID, err)
	}

	updatedGame, err := getGameWithPlayers(tx, gameID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing game update: %v", err)
	}
	committed = true

	return writeGameJSONResponse(w, http.StatusOK, updatedGame)
}

// KickGamePlayer removes a non-owner player from a setup-stage game.
//
// @Summary Kick a game player
// @Description This endpoint lets the authenticated game owner remove a non-owner player from a setup-stage game. Kicked players lose access to game state endpoints immediately.
// @Tags game
// @Accept json
// @Produce json
// @Param body body KickGamePlayerArgs true "kick player request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/kick_player [post]
func KickGamePlayer(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := KickGamePlayerArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}
	if args.AccountId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "account_id must be specified")
	}

	session, err := requireGameSession(w, store, args)
	if err != nil {
		return err
	}

	tx, err := beginGameTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting game update: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	game, err := getGameByIDForUpdate(tx, args.GameId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("no game exists with the ID %d", args.GameId))
		}
		return err
	}
	if err := ensureGameOwner(w, session, game); err != nil {
		return err
	}
	if err := ensureGameSetupMutable(w, game); err != nil {
		return err
	}
	if game.OwnerAccountID != nil && *game.OwnerAccountID == args.AccountId {
		return writeGameStatusError(w, http.StatusForbidden, "the game owner cannot be kicked")
	}
	if _, err := getGamePlayerByAccount(tx, args.GameId, args.AccountId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("account %d is not joined to the game", args.AccountId))
		}
		return err
	}
	if _, err := tx.Exec(`DELETE FROM game_players WHERE game_id = $1 AND account_id = $2`, args.GameId, args.AccountId); err != nil {
		return fmt.Errorf("an error occurred while removing the game player from the database: %v", err)
	}

	updatedGame, err := getGameWithPlayers(tx, args.GameId)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing game update: %v", err)
	}
	committed = true

	return writeGameJSONResponse(w, http.StatusOK, updatedGame)
}

// SelectGameCountry updates the authenticated player's selected country before the game starts.
//
// @Summary Select a game country
// @Description This endpoint lets an authenticated joined player choose a unique country during setup. Duplicate country selections are rejected, and country changes are blocked once the game has started.
// @Tags game
// @Accept json
// @Produce json
// @Param body body SelectGameCountryArgs true "country selection request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/select_country [post]
func SelectGameCountry(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := SelectGameCountryArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}
	country := strings.TrimSpace(args.Country)
	if country == "" {
		return writeGameStatusError(w, http.StatusBadRequest, "country must be specified")
	}

	session, err := requireGameSession(w, store, args)
	if err != nil {
		return err
	}

	tx, err := beginGameTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting game update: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	game, err := getGameByIDForUpdate(tx, args.GameId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("no game exists with the ID %d", args.GameId))
		}
		return err
	}
	if err := ensureGameSetupMutable(w, game); err != nil {
		return err
	}
	if _, err := ensureGamePlayer(w, tx, args.GameId, int64(session.AccountID)); err != nil {
		return err
	}

	var duplicate struct {
		ID int64 `db:"id"`
	}
	if err := tx.Get(&duplicate, `SELECT id FROM game_players WHERE game_id = $1 AND account_id <> $2 AND lower(country) = lower($3)`, args.GameId, int64(session.AccountID), country); err == nil {
		return writeGameStatusError(w, http.StatusConflict, ErrorCountryAlreadyTaken)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("an error occurred while validating the selected country for the game with the ID %d: %v", args.GameId, err)
	}

	if _, err := tx.Exec(`UPDATE game_players SET country = $1, updated_at = NOW() WHERE game_id = $2 AND account_id = $3`, country, args.GameId, int64(session.AccountID)); err != nil {
		return fmt.Errorf("an error occurred while updating the game player in the database: %v", err)
	}

	updatedGame, err := getGameWithPlayers(tx, args.GameId)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing game update: %v", err)
	}
	committed = true

	return writeGameJSONResponse(w, http.StatusOK, updatedGame)
}

// UpdateGameStartingSettlers updates the setup-stage starting settlers value for the authenticated owner.
//
// @Summary Update starting settlers
// @Description This endpoint lets the authenticated game owner update the setup-stage starting settlers value before the game starts.
// @Tags game
// @Accept json
// @Produce json
// @Param body body UpdateGameStartingSettlersArgs true "starting settlers request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/update_starting_settlers [post]
func UpdateGameStartingSettlers(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := UpdateGameStartingSettlersArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}

	return updateStartingValue(w, db, store, args, args.GameId, args.StartingSettlers, ErrorNegativeSettlers, "starting_settlers")
}

// UpdateGameStartingGold updates the setup-stage starting gold value for the authenticated owner.
//
// @Summary Update starting gold
// @Description This endpoint lets the authenticated game owner update the setup-stage starting gold value before the game starts.
// @Tags game
// @Accept json
// @Produce json
// @Param body body UpdateGameStartingGoldArgs true "starting gold request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/update_starting_gold [post]
func UpdateGameStartingGold(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := UpdateGameStartingGoldArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}

	return updateStartingValue(w, db, store, args, args.GameId, args.StartingGold, ErrorNegativeGold, "starting_gold")
}

// UpdateGamePing stores the authenticated player's latest ping value.
//
// @Summary Update player ping
// @Description This endpoint lets an authenticated joined player store a non-negative ping value that is returned in subsequent game responses.
// @Tags game
// @Accept json
// @Produce json
// @Param body body UpdateGamePingArgs true "ping update request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/update_ping [post]
func UpdateGamePing(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := UpdateGamePingArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}

	session, err := requireGameSession(w, store, args)
	if err != nil {
		return err
	}
	if args.Ping < 0 {
		return writeGameStatusError(w, http.StatusBadRequest, ErrorNegativePing)
	}

	tx, err := beginGameTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting game update: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	game, err := getGameByID(tx, args.GameId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("no game exists with the ID %d", args.GameId))
		}
		return err
	}
	if _, err := ensureGamePlayer(w, tx, args.GameId, int64(session.AccountID)); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE game_players SET ping_ms = $1, updated_at = NOW() WHERE game_id = $2 AND account_id = $3`, args.Ping, args.GameId, int64(session.AccountID)); err != nil {
		return fmt.Errorf("an error occurred while updating the game player in the database: %v", err)
	}

	updatedGame, err := getGameWithPlayers(tx, game.ID)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing game update: %v", err)
	}
	committed = true

	return writeGameJSONResponse(w, http.StatusOK, updatedGame)
}

// StartGame transitions a setup-stage game into the started state.
//
// @Summary Start a game
// @Description This endpoint lets the authenticated game owner start a setup-stage game once it has at least one joined player and any lobby dependency is still available.
// @Tags game
// @Accept json
// @Produce json
// @Param body body StartGameArgs true "start game request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/start_game [post]
// StartGame starts a game using the production server-owned seed source.
func StartGame(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	return startGame(w, r, db, store, generateStartWorldSeed, defaultStartGameplaySources())
}

// StartGameWithSeed is an integration-test seam. The supplied source is server
// configuration, never request data, so clients cannot choose gameplay seeds.
func StartGameWithSeed(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore, seedSource func() (int64, error)) error {
	return StartGameWithSeedAndSources(w, r, db, store, seedSource, defaultStartGameplaySources())
}

// StartGameWithSeedAndSources injects only server-owned deterministic sources for integration tests.
func StartGameWithSeedAndSources(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore, seedSource func() (int64, error), sources engine.Sources) error {
	if seedSource == nil {
		seedSource = generateStartWorldSeed
	}
	return startGame(w, r, db, store, seedSource, sources)
}

func startGame(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore, seedSource func() (int64, error), sources engine.Sources) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := StartGameArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}

	session, err := requireGameSession(w, store, args)
	if err != nil {
		return err
	}

	tx, err := beginGameTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting game update: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	game, err := getGameByIDForUpdate(tx, args.GameId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("no game exists with the ID %d", args.GameId))
		}
		return err
	}
	if err := ensureGameOwner(w, session, game); err != nil {
		return err
	}
	if err := ValidateWorldStartContract(game, false); err != nil {
		return writeGameStatusError(w, http.StatusConflict, err.Error())
	}
	if game.LobbyID != nil {
		if err := validateLobbyDependency(tx, *game.LobbyID); err != nil {
			return writeGameStatusError(w, http.StatusConflict, err.Error())
		}
	}

	players, err := listGamePlayersByGameID(tx, args.GameId)
	if err != nil {
		return err
	}
	if len(players) == 0 {
		return writeGameStatusError(w, http.StatusConflict, "game must have at least one joined player before it can be started")
	}

	selectedCountries := make(map[string]struct{}, len(players))
	for _, player := range players {
		country := strings.ToLower(strings.TrimSpace(player.Country))
		if country == "" {
			continue
		}
		if _, exists := selectedCountries[country]; exists {
			return writeGameStatusError(w, http.StatusConflict, ErrorCountryAlreadyTaken)
		}
		selectedCountries[country] = struct{}{}
	}

	seed, err := seedSource()
	if err != nil {
		return fmt.Errorf("an error occurred while generating the initial world seed for game %d: %v", args.GameId, err)
	}
	generationInput, err := NormalizeWorldGenerationInput(game, players, seed)
	if err != nil {
		return fmt.Errorf("an error occurred while normalizing world generation input for game %d: %v", args.GameId, err)
	}
	snapshot, turn, err := world.GenerateWorld(generationInput)
	if err != nil {
		return fmt.Errorf("an error occurred while generating world for game %d: %v", args.GameId, err)
	}
	snapshot.Turn = turn.Clone()
	snapshot = ApplyWorldSpawnCountries(snapshot, players)
	if err := createStartedGameplay(tx, snapshot, game, players, sources); err != nil {
		if errors.Is(err, errStartingAdvanceSetup) {
			return writeGameStatusError(w, http.StatusBadRequest, errStartingAdvanceSetup.Error())
		}
		if errors.Is(err, world.ErrWorldAlreadyExists) {
			return writeGameStatusError(w, http.StatusConflict, err.Error())
		}
		return fmt.Errorf("an error occurred while storing the initial world for game %d: %v", args.GameId, err)
	}

	if _, err := tx.Exec(`UPDATE games SET status = $1, started_at = NOW(), updated_at = NOW() WHERE id = $2`, "started", args.GameId); err != nil {
		return fmt.Errorf("an error occurred while updating the game with the ID %d: %v", args.GameId, err)
	}

	updatedGame, err := getGameWithPlayers(tx, args.GameId)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing game update: %v", err)
	}
	committed = true

	return writeGameJSONResponse(w, http.StatusOK, updatedGame)
}
