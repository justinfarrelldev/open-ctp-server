package engine

import (
	"encoding/json"
	"errors"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

type setGovernmentPayload struct {
	GovernmentID string `json:"government_id"`
}

type setScienceRatePayload struct {
	SciencePercent int `json:"science_percent"`
}

type setSocialPolicyPayload struct {
	Workday int `json:"workday"`
	Wages   int `json:"wages"`
	Rations int `json:"rations"`
}

func applyGovernmentCommand(state *State, command Command, runtime *rules.RuntimeRules) (string, json.RawMessage, error) {
	if command.Type == CommandSetScienceRate {
		return applyScienceRateCommand(state, command)
	}
	if command.Type == CommandSetSocialPolicy {
		return applySocialPolicyCommand(state, command)
	}
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

func applySocialPolicyCommand(state *State, command Command) (string, json.RawMessage, error) {
	var payload setSocialPolicyPayload
	if err := json.Unmarshal(command.Payload, &payload); err != nil || payload.Workday < -10 || payload.Workday > 10 || payload.Wages < -10 || payload.Wages > 10 || payload.Rations < -10 || payload.Rations > 10 {
		return "", nil, errors.New("social policy payload is invalid")
	}
	player, ok := playerByID(state, command.ActorPlayerID)
	if !ok {
		return "", nil, errors.New("owner is not a game player")
	}
	nation := nationForID(state, player.NationID)
	if nation == nil {
		return "", nil, errors.New("government is unavailable")
	}
	workday, wages, rations := payload.Workday, payload.Wages, payload.Rations
	nation.Workday = &workday
	nation.Wages = &wages
	nation.Rations = &rations
	encoded, _ := json.Marshal(payload)
	return EventSocialPolicyChanged, encoded, nil
}

func applyScienceRateCommand(state *State, command Command) (string, json.RawMessage, error) {
	var payload setScienceRatePayload
	if err := json.Unmarshal(command.Payload, &payload); err != nil || payload.SciencePercent < 0 || payload.SciencePercent > 100 {
		return "", nil, errors.New("science rate payload is invalid")
	}
	player, ok := playerByID(state, command.ActorPlayerID)
	if !ok {
		return "", nil, errors.New("owner is not a game player")
	}
	nation := nationForID(state, player.NationID)
	if nation == nil {
		return "", nil, errors.New("government is unavailable")
	}
	science := payload.SciencePercent
	nation.SciencePercent = &science
	encoded, _ := json.Marshal(payload)
	return EventScienceRateChanged, encoded, nil
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
	if government.CrimeOffset > 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":crime-offset", Kind: EffectCrimeOffset, Value: government.CrimeOffset})
	}
	if government.CrimeCoef > 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":crime-coef", Kind: EffectCrimeCoef, Value: int(government.CrimeCoef*100 + 0.5)})
	}
	if government.MaxSciencePercent != nil {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":max-science", Kind: EffectMaxSciencePercent, Value: *government.MaxSciencePercent})
	}
	if government.MartialLawMaxUnits > 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":martial-law-max-units", Kind: EffectMartialLawMaxUnits, Value: government.MartialLawMaxUnits})
	}
	if government.MartialLawEffect > 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":martial-law-effect", Kind: EffectMartialLawEffect, Value: government.MartialLawEffect})
	}
	if government.MartialLawThreshold > 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":martial-law-threshold", Kind: EffectMartialLawThreshold, Value: government.MartialLawThreshold})
	}
	if government.TooManyCitiesThreshold > 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":too-many-cities-threshold", Kind: EffectTooManyCitiesThreshold, Value: government.TooManyCitiesThreshold})
	}
	if government.TooManyCitiesCoefficient > 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":too-many-cities-coefficient", Kind: EffectTooManyCitiesCoefficient, Value: int(government.TooManyCitiesCoefficient*100 + 0.5)})
	}
	if government.EmpireDistanceScale > 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":empire-distance-scale", Kind: EffectEmpireDistanceScale, Value: int(government.EmpireDistanceScale*1000000 + 0.5)})
	}
	if government.MinEmpireDistance > 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":min-empire-distance", Kind: EffectMinEmpireDistance, Value: government.MinEmpireDistance})
	}
	if government.MaxEmpireDistance > 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":max-empire-distance", Kind: EffectMaxEmpireDistance, Value: government.MaxEmpireDistance})
	}
	appendGovernmentSocialEffects(nation, government)
	if government.PublicWorksBonus > 0 {
		nation.PublicWorks += government.PublicWorksBonus
	}
}

func appendGovernmentSocialEffects(nation *Nation, government rules.RuntimeGovernmentDefinition) {
	if government.WorkdayExpectation != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":workday-expectation", Kind: EffectWorkdayExpectation, Value: government.WorkdayExpectation})
	}
	if government.PositiveWorkdayCoef != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":positive-workday-coef", Kind: EffectPositiveWorkdayCoef, Value: int(government.PositiveWorkdayCoef*100 + 0.5)})
	}
	if government.NegativeWorkdayCoef != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":negative-workday-coef", Kind: EffectNegativeWorkdayCoef, Value: int(government.NegativeWorkdayCoef*100 + 0.5)})
	}
	if government.WagesExpectation != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":wages-expectation", Kind: EffectWagesExpectation, Value: government.WagesExpectation})
	}
	if government.PositiveWagesCoef != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":positive-wages-coef", Kind: EffectPositiveWagesCoef, Value: int(government.PositiveWagesCoef*100 + 0.5)})
	}
	if government.NegativeWagesCoef != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":negative-wages-coef", Kind: EffectNegativeWagesCoef, Value: int(government.NegativeWagesCoef*100 + 0.5)})
	}
	if government.RationsExpectation != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":rations-expectation", Kind: EffectRationsExpectation, Value: government.RationsExpectation})
	}
	if government.PositiveRationsCoef != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":positive-rations-coef", Kind: EffectPositiveRationsCoef, Value: int(government.PositiveRationsCoef*100 + 0.5)})
	}
	if government.NegativeRationsCoef != 0 {
		nation.Effects = append(nation.Effects, Effect{ID: "government:" + government.ID + ":negative-rations-coef", Kind: EffectNegativeRationsCoef, Value: int(government.NegativeRationsCoef*100 + 0.5)})
	}
}
