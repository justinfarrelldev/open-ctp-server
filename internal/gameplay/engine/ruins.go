package engine

import (
	"errors"
	"math"
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// Ruin stores the original two private rolls, not a client-selected reward. The
// immutable profile versions the released risks.txt RISK_VILLAGES_ONLY rules.
// Barbarian risk profiles are deliberately not advertised or approximated.
type Ruin struct {
	Profile   string `json:"profile"`
	TypeRoll  int    `json:"type_roll"`
	ValueRoll int    `json:"value_roll"`
}

const VillageRuinProfile = "villages-only-v1"

func validateRuins(state State) error {
	for _, tile := range state.Tiles {
		if ruin := tile.Ruin; ruin != nil {
			if ruin.Profile != VillageRuinProfile || ruin.TypeRoll < 0 || ruin.TypeRoll >= 100 || ruin.ValueRoll < 0 || ruin.ValueRoll >= 10000 {
				return errors.New("invalid ruin rolls or profile")
			}
		}
	}
	return nil
}

// exploreRuin is called only after successful entry, never by visibility or a
// preview. Atomic moves validate the entire path before opening intermediate huts.
func exploreRuin(state *State, playerID, tileID string, runtime *rules.RuntimeRules) {
	if runtime == nil {
		return
	}
	index := -1
	for i := range state.Tiles {
		if state.Tiles[i].ID == tileID {
			index = i
			break
		}
	}
	if index < 0 || state.Tiles[index].Ruin == nil {
		return
	}
	tile := state.Tiles[index]
	if isWaterTerrain(EffectiveTerrain(*state, tileID)) {
		return
	}
	nation := nationForID(state, nationForPlayer(*state, playerID))
	if nation == nil {
		return
	}
	ruin := *tile.Ruin
	state.Tiles[index].Ruin = nil
	notice := Notification{ID: "ruin:" + tileID, PlayerID: playerID, Kind: "ruin.explored", Turn: state.Phase.Turn, TileID: tileID, RewardKind: "empty"}
	switch {
	case ruin.TypeRoll < 30:
		amount := 150 + 150*ruin.ValueRoll/10000
		if nation.Gold <= math.MaxInt-amount {
			nation.Gold += amount
			notice.Amount = amount
			notice.RewardKind = "gold"
		}
	case ruin.TypeRoll < 45:
		var candidates []rules.RuntimeAdvanceDefinition
		for _, advance := range runtime.Advances {
			if !containsString(nation.Advances, advance.ID) && !containsString(advance.Prerequisites, advance.ID) && !advanceBeyondEndAge(advance, *runtime) && missingRuinAdvances(*nation, runtime, advance.ID, 1, map[string]bool{}) <= 1 {
				candidates = append(candidates, advance)
			}
		}
		sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
		if len(candidates) > 0 {
			advance := candidates[len(candidates)*ruin.ValueRoll/10000]
			nation.Advances = append(nation.Advances, advance.ID)
			sort.Strings(nation.Advances)
			maybeEnterAdvanceAge(nation, advance, *runtime)
			applyAdvanceTriggerFeat(state, nation, advance, runtime)
			clearCompletedResearchGoal(nation)
			if nation.ResearchTargetID == advance.ID {
				nation.ResearchProgress = 0
				nation.ResearchTargetID = nextResearchAdvance(*nation, *runtime)
			}
			notice.RewardKind = "advance"
			notice.ContentID = advance.ID
		}
	case ruin.TypeRoll < 70 || (ruin.TypeRoll >= 75 && ruin.TypeRoll < 80):
		settler := ruin.TypeRoll >= 75
		var candidates []rules.RuntimeUnitDefinition
		for _, unit := range runtime.Units {
			effects := runtimeUnitEffects(unit)
			if unit.Domain != DomainLand || len(unitRequiredGovernmentIDs(unit)) > 0 || hasEffect(effects, EffectCantBuild) {
				continue
			}
			if settler {
				if !hasEffect(effects, EffectSettler) {
					continue
				}
			} else {
				if unit.Attack < 1 || hasEffect(effects, EffectHasPopAndCanBuild) {
					continue
				}
				needed := 0
				for _, id := range unit.Prerequisites {
					needed += missingRuinAdvances(*nation, runtime, id, 2, map[string]bool{})
					if needed > 2 {
						break
					}
				}
				if needed > 2 {
					continue
				}
			}
			candidates = append(candidates, unit)
		}
		sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
		if len(candidates) > 0 && stackCount(*state, tileID) < maxArmySize {
			choice := len(candidates) * ruin.ValueRoll / 10000
			if settler {
				choice = 0
				// Retail takes its first land settler record. The canonical runtime
				// sorts by ID, so use its explicit starter, not a future engineer.
				for index, candidate := range candidates {
					if candidate.ID == runtime.Starter.TypeID {
						choice = index
						break
					}
				}
			}
			definition := candidates[choice]
			id := "ruin-unit:" + tileID
			if _, exists := unitByID(*state, id); !exists {
				unit := Unit{ID: id, OwnerID: playerID, TileID: tileID}
				applyUnitDefinition(&unit, definition)
				unit.MaxHealth += ownerWonderEffectSum(*state, playerID, EffectIncreaseHP) + ownerEffectSum(*state, playerID, EffectIncreaseHP)
				if unit.MaxHealth < 1 {
					unit.MaxHealth = 1
				}
				unit.Health = unit.MaxHealth
				unit.Movement = refreshedUnitMovement(*state, unit, definition.Movement)
				state.Units = append(state.Units, unit)
				notice.RewardKind = "unit"
				notice.UnitID = id
				notice.ContentID = definition.ID
			}
		}
	case ruin.TypeRoll < 75:
		if _, exists := cityByTile(*state, tileID); !exists && (tile.OwnerID == "" || tile.OwnerID == playerID) {
			id := "ruin-city:" + tileID
			if _, exists := cityByID(*state, id); !exists {
				city := City{ID: id, Name: defaultCityName(*state, playerID, runtime), OwnerID: playerID, FounderID: playerID, TileID: tileID, Population: 1, Goods: cityGoodsInRadius(*state, tileID)}
				if starter, ok := runtime.Unit(runtime.Starter.TypeID); ok {
					city.TypeID = starter.SettleCityTypeID
				}
				seedStartingAgeCityBuildings(*state, &city, runtime)
				state.Cities = append(state.Cities, city)
				initializeCityTerritory(state, city, runtime)
				notice.RewardKind = "city"
				notice.CityID = id
			}
		}
	}
	appendNotification(state, notice)
}

// Like Advances::GetMinPrerequisites, counts unknown advances including the
// requested advance. Clamp to the eligibility limit and detect cycles instead of
// reproducing the original unbounded recursion/undersized unit scratch array.
func missingRuinAdvances(nation Nation, runtime *rules.RuntimeRules, id string, limit int, visiting map[string]bool) int {
	if containsString(nation.Advances, id) {
		return 0
	}
	if visiting[id] || len(visiting) > limit {
		return limit + 1
	}
	advance, ok := runtime.Advance(id)
	if !ok {
		return limit + 1
	}
	visiting[id] = true
	defer delete(visiting, id)
	needed := 1
	for _, prerequisite := range advance.Prerequisites {
		needed += missingRuinAdvances(nation, runtime, prerequisite, limit, visiting)
		if needed > limit {
			return limit + 1
		}
	}
	return needed
}
