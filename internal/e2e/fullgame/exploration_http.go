package fullgame

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/service"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type HTTPExplorationTransport struct {
	BaseURL           string
	Client            *http.Client
	GameID            int64
	SessionID         int64
	Cursor            int64
	Revision          int64
	Terminal          bool
	InitialProjection []byte
	Namespace         string
	nextCommand       int
}

type CombatTransition struct {
	Result     *engine.CombatResult
	Terminal   bool
	Projection []byte
}

func (t *HTTPExplorationTransport) Projection(context.Context) ([]byte, error) {
	if t == nil || len(t.InitialProjection) == 0 {
		return nil, errors.New("initial public projection is required")
	}
	return append([]byte(nil), t.InitialProjection...), nil
}

func (t *HTTPExplorationTransport) Bootstrap(ctx context.Context, cursor int64) (int64, []byte, error) {
	if t == nil || t.BaseURL == "" || t.GameID < 1 || t.SessionID < 1 || cursor < 0 {
		return 0, nil, errors.New("public exploration SSE bootstrap is invalid")
	}
	requestCtx, cancel := scopedRequestContext(ctx)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, fmt.Sprintf("%s/game/%d/events?session_id=%d", t.BaseURL, t.GameID, t.SessionID), nil)
	if err != nil {
		return 0, nil, err
	}
	if cursor < t.Cursor {
		return 0, nil, fmt.Errorf("public exploration SSE cursor regressed from %d to %d", t.Cursor, cursor)
	}
	req.Header.Set("Last-Event-ID", strconv.FormatInt(cursor, 10))
	req.Header.Set("X-Gameplay-Resync", "current")
	response, err := t.client().Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("public exploration SSE status %d", response.StatusCode)
	}
	var id int64
	var name, data string
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "id: "):
			id, err = strconv.ParseInt(strings.TrimPrefix(line, "id: "), 10, 64)
			if err != nil {
				return 0, nil, err
			}
		case strings.HasPrefix(line, "event: "):
			name = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			data = strings.TrimPrefix(line, "data: ")
		case line == "" && name == "resync":
			var frame struct {
				CurrentRevision int64           `json:"current_revision"`
				Snapshot        json.RawMessage `json:"snapshot"`
			}
			if err := json.Unmarshal([]byte(data), &frame); err != nil || frame.CurrentRevision < 1 {
				return 0, nil, errors.New("public exploration SSE resync projection is invalid")
			}
			var projection visibility.Projection
			if err := json.Unmarshal(frame.Snapshot, &projection); err != nil || projection.Revision != frame.CurrentRevision {
				return 0, nil, errors.New("public exploration SSE snapshot is invalid")
			}
			if id < t.Cursor || frame.CurrentRevision < t.Revision {
				return 0, nil, fmt.Errorf("public exploration SSE resync regressed cursor/revision from %d/%d to %d/%d", t.Cursor, t.Revision, id, frame.CurrentRevision)
			}
			if frame.CurrentRevision == t.Revision && len(t.InitialProjection) > 0 && !bytes.Equal(frame.Snapshot, t.InitialProjection) {
				return 0, nil, errors.New("public exploration SSE resync replaced the current revision with a different projection")
			}
			t.Cursor, t.Revision, t.InitialProjection = id, projection.Revision, append([]byte(nil), frame.Snapshot...)
			return t.Cursor, append([]byte(nil), t.InitialProjection...), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, nil, err
	}
	return 0, nil, errors.New("public exploration SSE did not provide a resync projection")
}

func (t *HTTPExplorationTransport) CurrentProjection(ctx context.Context) ([]byte, error) {
	if t == nil {
		return nil, errors.New("public exploration transport is required")
	}
	_, projection, err := t.Bootstrap(ctx, t.Cursor)
	return projection, err
}

func (t *HTTPExplorationTransport) Move(ctx context.Context, unitID string, path []string) ([]byte, error) {
	if t == nil || t.BaseURL == "" || t.GameID < 1 || t.SessionID < 1 || unitID == "" || len(path) != 1 {
		return nil, errors.New("public exploration move is invalid")
	}
	payload := struct {
		UnitID string   `json:"unit_id"`
		Path   []string `json:"path"`
	}{UnitID: unitID, Path: path}
	key, err := t.commandID("explore", payload)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(struct {
		SessionID      int64    `json:"session_id"`
		Expected       int64    `json:"expected_revision"`
		IdempotencyKey string   `json:"idempotency_key"`
		CommandID      string   `json:"command_id"`
		UnitID         string   `json:"unit_id"`
		Path           []string `json:"path"`
	}{SessionID: t.SessionID, Expected: t.Revision, IdempotencyKey: key, CommandID: key, UnitID: unitID, Path: path})
	if err != nil {
		return nil, err
	}
	requestCtx, cancel := scopedRequestContext(ctx)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, fmt.Sprintf("%s/game/%d/move", t.BaseURL, t.GameID), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := t.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("public exploration move status %d: %s", response.StatusCode, strings.TrimSpace(string(data)))
	}
	var result service.Result
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Revision != t.Revision+1 || len(result.ViewerProjection) == 0 {
		return nil, errors.New("public exploration move did not refresh projection")
	}
	t.Revision, t.InitialProjection = result.Revision, append([]byte(nil), result.ViewerProjection...)
	return append([]byte(nil), t.InitialProjection...), nil
}

func (t *HTTPExplorationTransport) Combat(ctx context.Context, attackerArmyID, defenderArmyID string) (CombatTransition, error) {
	if t == nil || t.BaseURL == "" || t.GameID < 1 || t.SessionID < 1 || attackerArmyID == "" || defenderArmyID == "" {
		return CombatTransition{}, errors.New("public exploration combat is invalid")
	}
	payload := struct {
		AttackerArmyID string `json:"attacker_army_id"`
		DefenderArmyID string `json:"defender_army_id"`
	}{attackerArmyID, defenderArmyID}
	key, err := t.commandID("combat", payload)
	if err != nil {
		return CombatTransition{}, err
	}
	body, err := json.Marshal(struct {
		SessionID      int64  `json:"session_id"`
		Expected       int64  `json:"expected_revision"`
		IdempotencyKey string `json:"idempotency_key"`
		CommandID      string `json:"command_id"`
		AttackerArmyID string `json:"attacker_army_id"`
		DefenderArmyID string `json:"defender_army_id"`
	}{t.SessionID, t.Revision, key, key, attackerArmyID, defenderArmyID})
	if err != nil {
		return CombatTransition{}, err
	}
	requestCtx, cancel := scopedRequestContext(ctx)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, fmt.Sprintf("%s/game/%d/combat", t.BaseURL, t.GameID), bytes.NewReader(body))
	if err != nil {
		return CombatTransition{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := t.client().Do(req)
	if err != nil {
		return CombatTransition{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return CombatTransition{}, fmt.Errorf("public exploration combat status %d", response.StatusCode)
	}
	var result service.Result
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return CombatTransition{}, err
	}
	if result.Revision != t.Revision+1 || len(result.ViewerProjection) == 0 || result.Combat == nil {
		return CombatTransition{}, errors.New("public exploration combat did not refresh a combat result and projection")
	}
	t.Revision, t.InitialProjection = result.Revision, append([]byte(nil), result.ViewerProjection...)
	return CombatTransition{Result: result.Combat, Terminal: result.Terminal, Projection: append([]byte(nil), t.InitialProjection...)}, nil
}

func (t *HTTPExplorationTransport) Command(ctx context.Context, name string, payload map[string]json.RawMessage) ([]byte, error) {
	if t == nil || t.BaseURL == "" || t.GameID < 1 || t.SessionID < 1 || !publicCommand(name) {
		return nil, errors.New("public gameplay command is invalid")
	}
	commandPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	key, err := t.commandID(name, commandPayload)
	if err != nil {
		return nil, err
	}
	envelope := make(map[string]json.RawMessage, len(payload)+4)
	for field, value := range payload {
		envelope[field] = append(json.RawMessage(nil), value...)
	}
	envelope["session_id"], _ = json.Marshal(t.SessionID)
	envelope["expected_revision"], _ = json.Marshal(t.Revision)
	envelope["idempotency_key"], _ = json.Marshal(key)
	envelope["command_id"], _ = json.Marshal(key)
	body, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	requestCtx, cancel := scopedRequestContext(ctx)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, fmt.Sprintf("%s/game/%d/%s", t.BaseURL, t.GameID, name), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := t.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("public gameplay command %s status %d at revision %d with receipt %s", name, response.StatusCode, t.Revision, key)
	}
	var result service.Result
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Revision != t.Revision+1 || len(result.ViewerProjection) == 0 {
		return nil, errors.New("public gameplay command did not refresh projection")
	}
	t.Revision, t.Terminal, t.InitialProjection = result.Revision, result.Terminal, append([]byte(nil), result.ViewerProjection...)
	return append([]byte(nil), t.InitialProjection...), nil
}

func (t *HTTPExplorationTransport) commandID(name string, payload any) (string, error) {
	canonical, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	prefix := t.Namespace
	if prefix == "" {
		prefix = "public"
	}
	digest := sha256.Sum256(canonical)
	return fmt.Sprintf("%s-%s-%d-%s", prefix, name, t.Revision, hex.EncodeToString(digest[:8])), nil
}

func (t *HTTPExplorationTransport) client() *http.Client {
	if t.Client != nil {
		return t.Client
	}
	return http.DefaultClient
}

func scopedRequestContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithCancel(parent)
}

func publicCommand(name string) bool {
	switch name {
	case "army-create", "pass", "found-city", "queue-production", "establish-embassy", "steal-technology",
		"set-research", "set-government", "propose-agreement", "respond-agreement", "embark", "disembark",
		"combat", "terraform", "build-improvement", "cancel-terraform", "cancel-production":
		return true
	default:
		return false
	}
}
