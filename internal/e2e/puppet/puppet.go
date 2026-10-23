// Package puppet selects human E2E actions from public bytes and submits them through public HTTP.
package puppet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/ai"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

type TestHarnessPolicyLevel string

const (
	TestHarnessPolicyLevelPassOnly        TestHarnessPolicyLevel = "test-harness-pass-only"
	TestHarnessPolicyLevelProductionFirst TestHarnessPolicyLevel = "test-harness-production-first"
)

type PlanningInput struct {
	ActorPlayerID          string
	ExpectedRevision       int64
	PublicProjectionJSON   []byte
	PublicHistoryJSON      []byte
	TestHarnessPolicyLevel TestHarnessPolicyLevel
}

type RawPlanningInput struct {
	PublicProjectionJSON []byte
	PublicHistoryJSON    []byte
}

// PublicExplorationStep is the serialized, actor-visible movement history accepted by the test harness.
type PublicExplorationStep struct {
	CommandType       string `json:"command_type"`
	DestinationTileID string `json:"destination_tile_id"`
	Sequence          int64  `json:"sequence"`
}

type HumanActor struct {
	GameID    int64
	SessionID int64
}

type HumanCommand struct {
	GameID           int64
	SessionID        int64
	ExpectedRevision int64
	IdempotencyKey   string
	CommandID        string
	Route            string
	Payload          json.RawMessage
}

type SubmissionResult struct {
	Revision int64
}

type HumanHTTPClient interface {
	SubmitHumanCommand(context.Context, HumanCommand) (SubmissionResult, error)
}

type Puppeteer struct {
	Client               HumanHTTPClient
	ObservePlanningInput func(RawPlanningInput)
}

type Result struct {
	Candidate ai.Candidate
	Revision  int64
}

func (p Puppeteer) SelectAndSubmit(ctx context.Context, actor HumanActor, input PlanningInput, idempotencyKey string) (Result, error) {
	if p.Client == nil || actor.GameID < 1 || actor.SessionID < 1 || input.ActorPlayerID == "" || input.ExpectedRevision < 1 || strings.TrimSpace(idempotencyKey) == "" {
		return Result{}, errors.New("puppet request is invalid")
	}
	if !json.Valid(input.PublicProjectionJSON) || !json.Valid(input.PublicHistoryJSON) {
		return Result{}, errors.New("puppet public planning bytes are invalid")
	}
	history, err := decodePublicHistory(input.PublicHistoryJSON)
	if err != nil {
		return Result{}, err
	}
	projection := cloneBytes(input.PublicProjectionJSON)
	publicHistory := cloneBytes(input.PublicHistoryJSON)
	if p.ObservePlanningInput != nil {
		p.ObservePlanningInput(RawPlanningInput{PublicProjectionJSON: cloneBytes(projection), PublicHistoryJSON: cloneBytes(publicHistory)})
	}
	candidates, err := ai.PlanWithHistory(projection, input.ActorPlayerID, history)
	if err != nil {
		return Result{}, fmt.Errorf("plan puppet command: %w", err)
	}
	selected, err := selectCandidate(candidates, input.TestHarnessPolicyLevel)
	if err != nil {
		return Result{}, err
	}
	route, err := publicRoute(selected.Type)
	if err != nil {
		return Result{}, err
	}
	submitted, err := p.Client.SubmitHumanCommand(ctx, HumanCommand{
		GameID: actor.GameID, SessionID: actor.SessionID, ExpectedRevision: input.ExpectedRevision,
		IdempotencyKey: idempotencyKey, CommandID: idempotencyKey, Route: route, Payload: cloneBytes(selected.Payload),
	})
	if err != nil {
		return Result{}, err
	}
	return Result{Candidate: selected, Revision: submitted.Revision}, nil
}

func decodePublicHistory(data []byte) ([]ai.ExplorationStep, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var public []PublicExplorationStep
	if err := decoder.Decode(&public); err != nil {
		return nil, errors.New("puppet public history must be a JSON array of exploration steps")
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return nil, errors.New("puppet public history must contain one JSON value")
	}
	history := make([]ai.ExplorationStep, len(public))
	var previousSequence int64
	for index, step := range public {
		if step.CommandType != engine.CommandMoveUnit || step.DestinationTileID == "" || step.Sequence <= previousSequence {
			return nil, errors.New("puppet public exploration history is invalid")
		}
		history[index] = ai.ExplorationStep{CommandType: step.CommandType, DestinationTileID: step.DestinationTileID, Sequence: step.Sequence}
		previousSequence = step.Sequence
	}
	return history, nil
}

func selectCandidate(candidates []ai.Candidate, level TestHarnessPolicyLevel) (ai.Candidate, error) {
	if len(candidates) == 0 {
		return ai.Candidate{}, errors.New("production AI planner returned no candidates")
	}
	switch level {
	case TestHarnessPolicyLevelProductionFirst:
		return candidates[0], nil
	case TestHarnessPolicyLevelPassOnly:
		for _, candidate := range candidates {
			if candidate.Type == engine.CommandPass {
				return candidate, nil
			}
		}
		return ai.Candidate{}, errors.New("production AI planner returned no pass candidate")
	default:
		return ai.Candidate{}, fmt.Errorf("unknown test harness policy level %q", level)
	}
}

func publicRoute(commandType string) (string, error) {
	routes := map[string]string{
		engine.CommandPass: "pass", engine.CommandRoundComplete: "round-complete", engine.CommandMoveUnit: "move",
		engine.CommandSetResearch: "set-research", engine.CommandSetGovernment: "set-government",
		engine.CommandProposeAgreement: "propose-agreement", engine.CommandRespondAgreement: "respond-agreement",
		engine.CommandEmbark: "embark", engine.CommandDisembark: "disembark", engine.CommandCombat: "combat",
		engine.CommandTerraform: "terraform", engine.CommandBuildImprovement: "build-improvement", engine.CommandCancelTerraform: "cancel-terraform",
		engine.CommandQueueProduction: "queue-production", engine.CommandCancelProduction: "cancel-production", engine.CommandRushBuy: "rush-buy",
		engine.CommandFortify: "fortify", engine.CommandFoundCity: "found-city", engine.CommandArmyCreate: "army-create",
		engine.CommandEstablishEmbassy: "establish-embassy",
	}
	route, ok := routes[commandType]
	if !ok {
		return "", fmt.Errorf("production AI command %q has no public human route", commandType)
	}
	return route, nil
}

func cloneBytes(value []byte) []byte {
	if value == nil {
		return nil
	}
	return append([]byte(nil), value...)
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type HTTPError struct {
	Status int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("public human command returned status %d", e.Status)
}

func (c Client) SubmitHumanCommand(ctx context.Context, command HumanCommand) (SubmissionResult, error) {
	if command.GameID < 1 || command.SessionID < 1 || command.ExpectedRevision < 1 || command.Route == "" || command.IdempotencyKey == "" || command.CommandID == "" {
		return SubmissionResult{}, errors.New("public human command is invalid")
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(command.Payload, &body); err != nil {
		return SubmissionResult{}, fmt.Errorf("decode public human payload: %w", err)
	}
	add := func(name string, value any) error {
		raw, err := json.Marshal(value)
		if err == nil {
			body[name] = raw
		}
		return err
	}
	if err := add("session_id", command.SessionID); err != nil {
		return SubmissionResult{}, err
	}
	if err := add("expected_revision", command.ExpectedRevision); err != nil {
		return SubmissionResult{}, err
	}
	if err := add("idempotency_key", command.IdempotencyKey); err != nil {
		return SubmissionResult{}, err
	}
	if err := add("command_id", command.CommandID); err != nil {
		return SubmissionResult{}, err
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return SubmissionResult{}, err
	}
	endpoint, err := url.JoinPath(strings.TrimRight(c.BaseURL, "/"), "game", strconv.FormatInt(command.GameID, 10), command.Route)
	if err != nil {
		return SubmissionResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return SubmissionResult{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return SubmissionResult{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, response.Body)
		return SubmissionResult{}, &HTTPError{Status: response.StatusCode}
	}
	var result SubmissionResult
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return SubmissionResult{}, err
	}
	return result, nil
}
