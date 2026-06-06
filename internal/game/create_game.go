package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

const (
	// CTP2 defaults for Task 9 are grounded in the repo's CTP2 references:
	// - ctp2/ctp2_source_code/ctp2_code/gs/utility/gameinit.cpp
	// - ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp
	// - ctp2/ctp2_source_code/ctp2_code/gs/database/DiffDB.cpp
	// - ctp2/ctp2_released_game/ctp2_data/default/gamedata/DiffDB.txt
	// For the user-facing server setup baseline we keep the human-created game defaults at
	// 1 settler and 100 starting gold, with Task 11+ left to handle AI/difficulty-specific variance.
	DefaultGameStyle           = "standard_ctp2"
	DefaultGameStatus          = "setup"
	DefaultStartingSettlers    = 1
	DefaultStartingGold        = 100
	defaultRulesJSON           = "[]"
	defaultExclusionsJSON      = "{}"
	ErrorPasswordTooShort      = "password must be longer than 6 characters"
	ErrorPasswordRequired      = "password is required when password_protected is true"
	ErrorPasswordUnexpected    = "password was provided despite password_protected being set to false"
	ErrorNegativeSettlers      = "starting_settlers must be greater than or equal to 0"
	ErrorNegativeGold          = "starting_gold must be greater than or equal to 0"
	ErrorRulesJSONInvalid      = "rules_json must be a valid JSON array"
	ErrorExclusionsJSONInvalid = "exclusions_json must be a valid JSON object"
)

const insertGameQuery = `INSERT INTO games (
	name,
	owner_account_id,
	lobby_id,
	password_protected,
	password,
	is_locked,
	is_public,
	style,
	rules_json,
	exclusions_json,
	status,
	starting_settlers,
	starting_gold
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING id`

// CreateGameArgs represents the expected structure of the request body for creating a game.
//
// @Description Structure for the game creation request payload.
type CreateGameArgs struct {
	SessionId         *int64          `json:"session_id,omitempty"`
	Name              string          `json:"name"`
	OwnerAccountID    *int64          `json:"owner_account_id,omitempty"`
	LobbyID           *int64          `json:"lobby_id,omitempty"`
	PasswordProtected bool            `json:"password_protected"`
	Password          string          `json:"password"`
	IsLocked          *bool           `json:"is_locked,omitempty"`
	IsPublic          *bool           `json:"is_public,omitempty"`
	Style             *string         `json:"style,omitempty"`
	RulesJSON         json.RawMessage `json:"rules_json,omitempty"`
	ExclusionsJSON    json.RawMessage `json:"exclusions_json,omitempty"`
	StartingSettlers  *int            `json:"starting_settlers,omitempty"`
	StartingGold      *int            `json:"starting_gold,omitempty"`
}

func (a CreateGameArgs) GetSessionID() *int64 {
	return a.SessionId
}

type createGameRecord struct {
	Name              string
	OwnerAccountID    int64
	LobbyID           *int64
	PasswordProtected bool
	Password          string
	IsLocked          bool
	IsPublic          bool
	Style             string
	RulesJSON         string
	ExclusionsJSON    string
	Status            string
	StartingSettlers  int
	StartingGold      int
}

func decodeCreateGameArgs(r *http.Request) (CreateGameArgs, error) {
	args := CreateGameArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return CreateGameArgs{}, err
	}

	return args, nil
}

func resolveCreateGameRecord(args CreateGameArgs, session *auth.Session) (createGameRecord, error) {
	record := createGameRecord{
		Name:              args.Name,
		OwnerAccountID:    int64(session.AccountID),
		LobbyID:           args.LobbyID,
		PasswordProtected: args.PasswordProtected,
		Password:          args.Password,
		IsLocked:          false,
		IsPublic:          false,
		Style:             DefaultGameStyle,
		RulesJSON:         defaultRulesJSON,
		ExclusionsJSON:    defaultExclusionsJSON,
		Status:            DefaultGameStatus,
		StartingSettlers:  DefaultStartingSettlers,
		StartingGold:      DefaultStartingGold,
	}

	if args.PasswordProtected && strings.TrimSpace(args.Password) == "" {
		return createGameRecord{}, fmt.Errorf(ErrorPasswordRequired)
	}
	if !args.PasswordProtected && args.Password != "" {
		return createGameRecord{}, fmt.Errorf(ErrorPasswordUnexpected)
	}
	if args.Password != "" && len(args.Password) < 6 {
		return createGameRecord{}, fmt.Errorf(ErrorPasswordTooShort)
	}
	if args.IsLocked != nil {
		record.IsLocked = *args.IsLocked
	}
	if args.IsPublic != nil {
		record.IsPublic = *args.IsPublic
	}
	if args.Style != nil && strings.TrimSpace(*args.Style) != "" {
		record.Style = strings.TrimSpace(*args.Style)
	}
	if args.StartingSettlers != nil {
		if *args.StartingSettlers < 0 {
			return createGameRecord{}, fmt.Errorf(ErrorNegativeSettlers)
		}
		record.StartingSettlers = *args.StartingSettlers
	}
	if args.StartingGold != nil {
		if *args.StartingGold < 0 {
			return createGameRecord{}, fmt.Errorf(ErrorNegativeGold)
		}
		record.StartingGold = *args.StartingGold
	}

	rulesJSON, err := normalizeJSONArray(args.RulesJSON, []byte(defaultRulesJSON))
	if err != nil {
		return createGameRecord{}, fmt.Errorf(ErrorRulesJSONInvalid)
	}
	record.RulesJSON = rulesJSON

	exclusionsJSON, err := normalizeJSONObject(args.ExclusionsJSON, []byte(defaultExclusionsJSON))
	if err != nil {
		return createGameRecord{}, fmt.Errorf(ErrorExclusionsJSONInvalid)
	}
	record.ExclusionsJSON = exclusionsJSON

	return record, nil
}

func normalizeJSONArray(raw json.RawMessage, defaultValue []byte) (string, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return string(defaultValue), nil
	}

	var decoded []any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return "", err
	}

	normalized, err := json.Marshal(decoded)
	if err != nil {
		return "", err
	}

	return string(normalized), nil
}

func normalizeJSONObject(raw json.RawMessage, defaultValue []byte) (string, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return string(defaultValue), nil
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return "", err
	}

	normalized, err := json.Marshal(decoded)
	if err != nil {
		return "", err
	}

	return string(normalized), nil
}

// CreateGame creates a new game owned by the authenticated session account.
//
// @Summary Create a new game
// @Description This endpoint creates a new multiplayer game setup record using CTP2-first defaults for omitted setup fields.
// @Tags game
// @Accept json
// @Produce json
// @Param body body CreateGameArgs true "Game creation request body"
// @Success 201 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/create_game [post]
func CreateGame(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args, err := decodeCreateGameArgs(r)
	if err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}

	session, err := requireGameSession(w, store, args)
	if err != nil {
		return err
	}

	record, err := resolveCreateGameRecord(args, session)
	if err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, err.Error())
	}

	tx, err := beginGameTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting game creation: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	gamePlayerAccountIDs, err := resolveGamePlayerAccountIDs(tx, record.OwnerAccountID, record.LobbyID)
	if err != nil {
		return err
	}

	var lobbyIDArg any
	if record.LobbyID != nil {
		lobbyIDArg = *record.LobbyID
	}

	var gameID int64
	if err := tx.QueryRowx(
		insertGameQuery,
		record.Name,
		record.OwnerAccountID,
		lobbyIDArg,
		record.PasswordProtected,
		record.Password,
		record.IsLocked,
		record.IsPublic,
		record.Style,
		record.RulesJSON,
		record.ExclusionsJSON,
		record.Status,
		record.StartingSettlers,
		record.StartingGold,
	).Scan(&gameID); err != nil {
		return fmt.Errorf("an error occurred while storing the game in the database: %v", err)
	}

	if err := seedGamePlayers(tx, gameID, gamePlayerAccountIDs); err != nil {
		return err
	}

	game, err := getGameWithPlayers(tx, gameID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing game creation: %v", err)
	}
	committed = true

	return writeGameJSONResponse(w, http.StatusCreated, game)
}
