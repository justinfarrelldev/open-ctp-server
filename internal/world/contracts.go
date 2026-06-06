package world

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

var ErrWorldContractsUnimplemented = errors.New("world contracts are not implemented yet")

const SpaceLayerUnsupportedError = "space layer is not supported for CTP2 worlds"

type GenerationHostParams struct {
	Style             string `json:"style"`
	RulesJSON         string `json:"rules_json"`
	ExclusionsJSON    string `json:"exclusions_json"`
	StartingSettlers  int    `json:"starting_settlers"`
	StartingGold      int    `json:"starting_gold"`
	Seed              *int64 `json:"seed,omitempty"`
	Wetness           *int    `json:"wetness,omitempty"`
	MapSize           string `json:"map_size,omitempty"`
	Climate           string `json:"climate,omitempty"`
	GeneratorVersion  string `json:"generator_version,omitempty"`
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

func GetWorldHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	if strings.TrimSpace(r.PathValue("game_id")) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("game_id must be specified"))
		return
	}

	http.Error(w, ErrWorldContractsUnimplemented.Error(), http.StatusNotImplemented)
}

func GetWorldSpaceLayerHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	if strings.TrimSpace(r.PathValue("game_id")) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("game_id must be specified"))
		return
	}

	http.Error(w, ErrWorldContractsUnimplemented.Error(), http.StatusNotImplemented)
}
