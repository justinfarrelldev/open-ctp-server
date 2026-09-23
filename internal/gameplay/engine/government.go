package engine

import (
	"encoding/json"
	"errors"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

type setGovernmentPayload struct {
	GovernmentID string `json:"government_id"`
}

func applyGovernmentCommand(state *State, command Command, runtime *rules.RuntimeRules) (string, json.RawMessage, error) {
	var payload setGovernmentPayload
	if err := json.Unmarshal(command.Payload, &payload); err != nil || payload.GovernmentID == "" {
		return "", nil, errors.New("government payload is invalid")
	}
	player, ok := playerByID(state, command.ActorPlayerID)
	if !ok {
		return "", nil, errors.New("owner is not a game player")
	}
	nation := nationForID(state, player.NationID)
	if nation == nil || runtime == nil {
		return "", nil, errors.New("government is unavailable")
	}
	government, ok := runtime.Government(payload.GovernmentID)
	if !ok {
		return "", nil, errors.New("government is unavailable")
	}
	if government.EnableAdvanceID != "" && !containsString(nation.Advances, government.EnableAdvanceID) {
		return "", nil, errors.New("government prerequisite is not met")
	}
	nation.GovernmentID = government.ID
	upsertGovernmentEffects(nation, government)
	encoded, _ := json.Marshal(payload)
	return EventGovernmentChanged, encoded, nil
}

func upsertGovernmentEffects(nation *Nation, government rules.RuntimeGovernmentDefinition) {
	filtered := nation.Effects[:0]
	for _, effect := range nation.Effects {
		if len(effect.ID) >= len("government:") && effect.ID[:len("government:")] == "government:" {
			continue
		}
		filtered = append(filtered, effect)
	}
	nation.Effects = filtered
	if government.ResearchBonus > 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":research", Kind: EffectResearch, Value: government.ResearchBonus})
	}
	if government.ResearchPercent != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":research-percent", Kind: EffectResearchPercent, Value: government.ResearchPercent})
	}
	if government.FoodPercent != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":food", Kind: EffectYieldFoodPercent, Value: government.FoodPercent})
	}
	if government.ProductionPercent != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":production", Kind: EffectYieldProductionPercent, Value: government.ProductionPercent})
	}
	if government.GoldPercent != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":gold", Kind: EffectYieldGoldPercent, Value: government.GoldPercent})
	}
	if government.PollutionPercent != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":pollution", Kind: EffectPollutionPercent, Value: government.PollutionPercent})
	}
	if government.PublicWorksBonus > 0 {
		nation.PublicWorks += government.PublicWorksBonus
	}
}
