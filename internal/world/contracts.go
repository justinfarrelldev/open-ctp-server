package world

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

var ErrWorldContractsUnimplemented = errors.New("world contracts are not implemented yet")

const SpaceLayerUnsupportedError = "space layer is not supported for CTP2 worlds"
const errorGamePlayerRequired = "only joined game players may perform this action"

type worldReadContractState string

const (
	worldReadContractStateReady     worldReadContractState = "ready"
	worldReadContractStateNonMember worldReadContractState = "non_member"
	worldReadContractStateSetupGame worldReadContractState = "setup_game"
	worldReadContractStateMissing   worldReadContractState = "missing_world"
)

type worldReadContractStateContextKey struct{}

type GenerationHostParams struct {
	Style            string `json:"style"`
	RulesJSON        string `json:"rules_json"`
	ExclusionsJSON   string `json:"exclusions_json"`
	StartingSettlers int    `json:"starting_settlers"`
	StartingGold     int    `json:"starting_gold"`
	Seed             *int64 `json:"seed,omitempty"`
	Wetness          *int   `json:"wetness,omitempty"`
	MapSize          string `json:"map_size,omitempty"`
	Climate          string `json:"climate,omitempty"`
	GeneratorVersion string `json:"generator_version,omitempty"`
}

type GenerationInput struct {
	GameID           int64                `json:"game_id"`
	PlayerOrder      []int64              `json:"player_order"`
	GeneratorVersion string               `json:"generator_version"`
	HostParams       GenerationHostParams `json:"host_params"`
}

type WorldTile struct {
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Terrain string `json:"terrain"`
}

type PlayerSpawn struct {
	AccountID int64  `json:"account_id"`
	Country   string `json:"country,omitempty"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
}

type WorldReadResponse struct {
	GameID           int64                `json:"game_id"`
	GameStatus       string               `json:"game_status"`
	GeneratorVersion string               `json:"generator_version"`
	HostParams       GenerationHostParams `json:"host_params"`
	PlayerOrder      []int64              `json:"player_order"`
	Spawns           []PlayerSpawn        `json:"spawns"`
	Tiles            []WorldTile          `json:"tiles"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func withWorldReadContractState(r *http.Request, state worldReadContractState) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), worldReadContractStateContextKey{}, state))
}

func getWorldReadContractState(r *http.Request) worldReadContractState {
	state, _ := r.Context().Value(worldReadContractStateContextKey{}).(worldReadContractState)
	return state
}

func writeWorldJSONResponse(w http.ResponseWriter, status int, payload any) {
	responseBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "an error occurred while encoding the response body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(responseBytes)
}

func parseWorldGameID(r *http.Request) (int64, error) {
	gameIDStr := strings.TrimSpace(r.PathValue("game_id"))
	if gameIDStr == "" {
		return 0, errors.New("game_id must be specified")
	}

	gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
	if err != nil || gameID <= 0 {
		return 0, errors.New("invalid game_id")
	}

	return gameID, nil
}

func GetWorldHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	gameID, err := parseWorldGameID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if _, err := auth.SessionIDFromQuery(r); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	switch getWorldReadContractState(r) {
	case worldReadContractStateNonMember:
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(errorGamePlayerRequired))
		return
	case worldReadContractStateSetupGame:
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(fmt.Sprintf("game %d has not started", gameID)))
		return
	case worldReadContractStateMissing:
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(fmt.Sprintf("world for game %d was not found", gameID)))
		return
	case worldReadContractStateReady:
		response, err := deterministicContractWorld(gameID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeWorldJSONResponse(w, http.StatusOK, response)
		return
	default:
		http.Error(w, ErrWorldContractsUnimplemented.Error(), http.StatusNotImplemented)
	}
}

func GetWorldSpaceLayerHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	if _, err := parseWorldGameID(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	writeWorldJSONResponse(w, http.StatusNotFound, ErrorResponse{Error: SpaceLayerUnsupportedError})
}

func deterministicContractWorld(gameID int64) (WorldReadResponse, error) {
	seed := int64(41041)
	wetness := 50
	snapshot, _, err := GenerateWorld(GenerationInput{
		GameID:           gameID,
		PlayerOrder:      []int64{101, 202, 303},
		GeneratorVersion: "ctp2-first-v1",
		HostParams: GenerationHostParams{
			Style:            "standard_ctp2",
			RulesJSON:        "[]",
			ExclusionsJSON:   "{}",
			StartingSettlers: 1,
			StartingGold:     100,
			Seed:             &seed,
			Wetness:          &wetness,
			MapSize:          "small",
			Climate:          "temperate",
			GeneratorVersion: "ctp2-first-v1",
		},
	})
	if err != nil {
		return WorldReadResponse{}, err
	}
	return snapshotToReadResponse(snapshot), nil
}
