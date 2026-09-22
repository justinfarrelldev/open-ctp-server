package fullgame

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type ConquestTemplate struct {
	ID              string
	AttackerArmyID  string
	RestartAtAction int
}

type ConquestObservation struct {
	First  Artifacts
	Second Artifacts
}

type PublicConquestExecutor interface {
	ExecutePublicConquest(context.Context, ConquestTemplate, string) (Artifacts, error)
}

func FixedConquestTemplate() ConquestTemplate {
	return ConquestTemplate{ID: "public-conquest-v1", AttackerArmyID: "attacker-army", RestartAtAction: 2}
}

func (template ConquestTemplate) Schedule(projection []byte, attackerRole, attackerUnitID, targetTileID, defenderArmyID string) ([]ScheduledStep, error) {
	if template.ID == "" || template.AttackerArmyID == "" || template.RestartAtAction < 1 || attackerRole == "" || attackerUnitID == "" || targetTileID == "" || defenderArmyID == "" {
		return nil, errors.New("fixed conquest template is invalid")
	}
	path, err := PlanVisibleRoute(projection, attackerUnitID, targetTileID)
	if err != nil {
		return nil, fmt.Errorf("plan visible conquest route: %w", err)
	}
	armyPayload, err := json.Marshal(struct {
		ArmyID  string   `json:"army_id"`
		UnitIDs []string `json:"unit_ids"`
	}{ArmyID: template.AttackerArmyID, UnitIDs: []string{attackerUnitID}})
	if err != nil {
		return nil, err
	}
	movePayload, err := json.Marshal(struct {
		Path   []string `json:"path"`
		UnitID string   `json:"unit_id"`
	}{Path: path, UnitID: attackerUnitID})
	if err != nil {
		return nil, err
	}
	combatPayload, err := json.Marshal(struct {
		AttackerArmyID string `json:"attacker_army_id"`
		DefenderArmyID string `json:"defender_army_id"`
	}{AttackerArmyID: template.AttackerArmyID, DefenderArmyID: defenderArmyID})
	if err != nil {
		return nil, err
	}
	return []ScheduledStep{
		{Role: attackerRole, Command: "army-create", Payload: armyPayload},
		{Role: attackerRole, Command: "move", Payload: movePayload},
		{Role: attackerRole, Command: "combat", Payload: combatPayload},
		{Role: attackerRole, Command: "pass", Payload: json.RawMessage(`{}`)},
	}, nil
}

func ObserveConquestTwice(ctx context.Context, template ConquestTemplate, executor PublicConquestExecutor) (ConquestObservation, error) {
	if executor == nil || template.ID == "" {
		return ConquestObservation{}, errors.New("public conquest observation is invalid")
	}
	first, err := executor.ExecutePublicConquest(ctx, template, "run-1")
	if err != nil {
		return ConquestObservation{}, err
	}
	second, err := executor.ExecutePublicConquest(ctx, template, "run-2")
	if err != nil {
		return ConquestObservation{}, err
	}
	if first != second {
		return ConquestObservation{}, fmt.Errorf("public conquest observation mismatch: first=%+v second=%+v", first, second)
	}
	if first.TerminalWinner == "" || first.TerminalReason != "conquest" || first.StateHash == "" || first.EventHash == "" {
		return ConquestObservation{}, errors.New("natural conquest artifacts were not observed")
	}
	return ConquestObservation{First: first, Second: second}, nil
}
