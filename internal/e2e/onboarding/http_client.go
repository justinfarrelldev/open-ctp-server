package onboarding

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/game"
)

// HTTPClient runs onboarding scripts against the service's public HTTP routes.
type HTTPClient struct {
	BaseURL string
	Client  *http.Client
	Clean   func(context.Context, string) error
}

func (c HTTPClient) Signup(ctx context.Context, name, email string) (Actor, error) {
	var session int64
	err := c.request(ctx, http.MethodPost, "/account/create_account", map[string]any{
		"account":  map[string]any{"name": name, "info": "", "location": "", "email": email, "experience_level": 1},
		"password": "onboarding-password",
	}, &session)
	if err != nil {
		return Actor{}, err
	}
	return Actor{Name: name, Email: email, Session: session}, nil
}

func (c HTTPClient) CreateLobby(ctx context.Context, actor Actor, name string) (int64, error) {
	var lobby struct {
		ID int64 `json:"id"`
	}
	err := c.request(ctx, http.MethodPost, "/lobby/create_lobby", map[string]any{
		"session_id": actor.Session,
		"password":   "lobby-password",
		"lobby":      map[string]any{"name": name, "is_public": true, "is_muted": false},
	}, &lobby)
	if err != nil {
		return 0, err
	}
	if lobby.ID == 0 {
		return 0, fmt.Errorf("create lobby returned no public identity")
	}
	return lobby.ID, nil
}

func (c HTTPClient) JoinLobby(ctx context.Context, actor Actor, lobbyID int64) error {
	return c.request(ctx, http.MethodPost, "/lobby/join_lobby", map[string]any{"session_id": actor.Session, "lobby_id": lobbyID}, nil)
}

func (c HTTPClient) OpenLobbyStream(ctx context.Context, actor Actor, lobbyID int64) (func() error, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(fmt.Sprintf("/lobby/connect_lobby?session_id=%d&lobby_id=%d", actor.Session, lobbyID)), nil)
	if err != nil {
		return nil, err
	}
	response, err := c.client().Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, statusError(response.StatusCode)
	}
	if response.Header.Get("Content-Type") != "text/event-stream" {
		response.Body.Close()
		return nil, fmt.Errorf("lobby stream content type %q", response.Header.Get("Content-Type"))
	}
	line, err := bufio.NewReader(response.Body).ReadString('\n')
	if err != nil {
		response.Body.Close()
		return nil, fmt.Errorf("read lobby stream event: %w", err)
	}
	if strings.TrimSpace(line) != "event: player_connected" {
		response.Body.Close()
		return nil, fmt.Errorf("unexpected lobby stream event %q", strings.TrimSpace(line))
	}
	return response.Body.Close, nil
}

func (c HTTPClient) ReadGame(ctx context.Context, actor Actor, gameID int64) (game.Game, error) {
	var result game.Game
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(fmt.Sprintf("/game/get_game?game_id=%d&session_id=%d", gameID, actor.Session)), nil)
	if err != nil {
		return game.Game{}, err
	}
	response, err := c.client().Do(req)
	if err != nil {
		return game.Game{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return game.Game{}, statusError(response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return game.Game{}, err
	}
	return result, nil
}

func (c HTTPClient) CreateGame(ctx context.Context, actor Actor, lobbyID int64, name string) (int64, error) {
	body := map[string]any{"session_id": actor.Session, "name": name, "password_protected": false, "password": ""}
	if lobbyID > 0 {
		body["lobby_id"] = lobbyID
	}
	var game struct {
		ID int64 `json:"id"`
	}
	err := c.request(ctx, http.MethodPost, "/game/create_game", body, &game)
	return game.ID, err
}

func (c HTTPClient) ConfigureGame(ctx context.Context, actor Actor, gameID int64, setup GameSetup) error {
	if err := c.request(ctx, http.MethodPost, "/game/update_style", map[string]any{"session_id": actor.Session, "game_id": gameID, "style": setup.Style, "map_size": setup.MapSize, "climate": setup.Climate, "wetness": setup.Wetness}, nil); err != nil {
		return err
	}
	if setup.RulesJSON == "" {
		return nil
	}
	return c.request(ctx, http.MethodPost, "/game/update_rules", map[string]any{"session_id": actor.Session, "game_id": gameID, "rules_json": json.RawMessage(setup.RulesJSON)}, nil)
}

func (c HTTPClient) UpdateStartingSettlers(ctx context.Context, actor Actor, gameID int64, settlers int) error {
	return c.request(ctx, http.MethodPost, "/game/update_starting_settlers", map[string]any{"session_id": actor.Session, "game_id": gameID, "starting_settlers": settlers}, nil)
}

func (c HTTPClient) SelectCountry(ctx context.Context, actor Actor, gameID int64, country string) error {
	return c.request(ctx, http.MethodPost, "/game/select_country", map[string]any{"session_id": actor.Session, "game_id": gameID, "country": country}, nil)
}

func (c HTTPClient) AddAI(ctx context.Context, owner Actor, gameID, targetAccountID int64) error {
	return c.request(ctx, http.MethodPost, "/game/add_ai", map[string]any{"session_id": owner.Session, "game_id": gameID, "target_account_id": targetAccountID}, nil)
}

func (c HTTPClient) AddManagedAI(ctx context.Context, owner Actor, gameID int64, country, name string) error {
	return c.request(ctx, http.MethodPost, "/game/add_ai", map[string]any{"session_id": owner.Session, "game_id": gameID, "country": country, "name": name}, nil)
}

func (c HTTPClient) StartGame(ctx context.Context, actor Actor, gameID int64) error {
	return c.request(ctx, http.MethodPost, "/game/start_game", map[string]any{"session_id": actor.Session, "game_id": gameID}, nil)
}

// GameplayCommand submits one documented public gameplay command and returns its revision.
func (c HTTPClient) GameplayCommand(ctx context.Context, actor Actor, gameID int64, command string, expectedRevision int64, idempotencyKey string, payload map[string]any) (int64, error) {
	body := make(map[string]any, len(payload)+4)
	for key, value := range payload {
		body[key] = value
	}
	body["session_id"] = actor.Session
	body["expected_revision"] = expectedRevision
	body["idempotency_key"] = idempotencyKey
	body["command_id"] = idempotencyKey
	var result struct {
		Revision int64 `json:"revision"`
	}
	if err := c.request(ctx, http.MethodPost, fmt.Sprintf("/game/%d/%s", gameID, command), body, &result); err != nil {
		return 0, err
	}
	return result.Revision, nil
}

// ReadProjectedWorld is a complete viewer-safe current read, never a download of
// generation inputs or the immutable canonical world snapshot.
func (c HTTPClient) ReadProjectedWorld(ctx context.Context, actor Actor, gameID int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(fmt.Sprintf("/game/%d/events?session_id=%d", gameID, actor.Session)), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Gameplay-Resync", "current")
	response, err := c.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, statusError(response.StatusCode)
	}
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	event, data := "", ""
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event:"):
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			if data != "" {
				data += "\n"
			}
			data += strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		case line == "":
			if event == "resync" {
				var envelope struct {
					Revision *int64          `json:"current_revision"`
					Snapshot json.RawMessage `json:"snapshot"`
				}
				if err := json.Unmarshal([]byte(data), &envelope); err != nil {
					return nil, err
				}
				var projection struct {
					Revision *int64 `json:"revision"`
				}
				if err := json.Unmarshal(envelope.Snapshot, &projection); err != nil {
					return nil, err
				}
				if envelope.Revision == nil || projection.Revision == nil || *envelope.Revision != *projection.Revision {
					return nil, fmt.Errorf("current resync revision mismatch")
				}
				return append([]byte(nil), envelope.Snapshot...), nil
			}
			event, data = "", ""
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("current resync did not return a viewer snapshot")
}

func (c HTTPClient) LeaveLobby(ctx context.Context, actor Actor, lobbyID int64) error {
	return c.request(ctx, http.MethodPost, "/lobby/leave_lobby", map[string]any{"session_id": actor.Session, "lobby_id": lobbyID}, nil)
}

func (c HTTPClient) AssertClean(ctx context.Context, namespace string) error {
	if c.Clean == nil {
		return fmt.Errorf("assert cleanup: no isolated-database verifier configured")
	}
	return c.Clean(ctx, namespace)
}

func (c HTTPClient) request(ctx context.Context, method, path string, body any, result any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint(path), bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := c.client().Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusBadRequest || response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusConflict {
		body, _ := io.ReadAll(response.Body)
		if len(body) != 0 {
			return fmt.Errorf("%w: %s", ErrConflict, strings.TrimSpace(string(body)))
		}
		return ErrConflict
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(response.Body)
		if len(body) != 0 {
			return fmt.Errorf("public endpoint status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
		}
		return statusError(response.StatusCode)
	}
	if result == nil {
		return nil
	}
	return json.NewDecoder(response.Body).Decode(result)
}

func (c HTTPClient) endpoint(path string) string { return strings.TrimRight(c.BaseURL, "/") + path }
func (c HTTPClient) client() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return http.DefaultClient
}
func statusError(status int) error { return fmt.Errorf("public endpoint status %d", status) }
