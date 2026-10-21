package engine

import (
	"errors"
	"math"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const CommandSetGovernor = "set-governor"

// GovernorState is owner-private city production automation. It is deliberately
// not the national slider optimizer or a second AI controller for the player.
type GovernorState struct {
	Enabled  bool   `json:"enabled"`
	Priority string `json:"priority"`
}

type SetGovernorPayload struct {
	CityID   string `json:"city_id"`
	Enabled  *bool  `json:"enabled"`
	Priority string `json:"priority,omitempty"`
}

var governorPriorities = []string{"growth", "production", "gold", "science", "happiness", "defense", "offense"}

// GovernorPriorities advertises only focuses with actual pinned retail lists.
func GovernorPriorities(runtime *rules.RuntimeRules) []string {
	if runtime == nil || runtime.AI == nil {
		return nil
	}
	var priorities []string
	for _, priority := range governorPriorities {
		for _, list := range append(append([]rules.RuntimeAIList(nil), runtime.AI.BuildingBuildLists...), runtime.AI.WonderBuildLists...) {
			if (list.ID == "BUILDING_BUILD_LIST_"+strings.ToUpper(priority) || list.ID == "WONDER_BUILD_LIST_"+strings.ToUpper(priority)) && len(list.Items) > 0 {
				priorities = append(priorities, priority)
				break
			}
		}
	}
	return priorities
}

func setGovernor(state *State, command Command, runtime *rules.RuntimeRules) (string, error) {
	if err := activeActor(*state, command.ActorPlayerID); err != nil {
		return "", err
	}
	var payload SetGovernorPayload
	if err := decodePayload(command.Payload, &payload); err != nil {
		return "", err
	}
	city, ok := cityByID(*state, payload.CityID)
	if !ok || city.OwnerID != command.ActorPlayerID || payload.Enabled == nil {
		return "", errors.New("governor command is invalid")
	}
	priority := payload.Priority
	if priority == "" && city.Governor != nil {
		priority = city.Governor.Priority
	}
	if *payload.Enabled || priority != "" {
		if !containsString(GovernorPriorities(runtime), priority) {
			return "", errors.New("governor priority is unsupported")
		}
	}
	city.Governor = &GovernorState{Enabled: *payload.Enabled, Priority: priority}
	if err := fillGovernorQueue(state, city, runtime); err != nil {
		return "", err
	}
	return "city.governor_changed", nil
}

// FillEmptyBuildQueues (governor.cpp) leaves human queues intact. This bounded
// implementation uses its ordered building/wonder lists and ordinary production
// legality. It does not emulate the original empire-wide desired army counts,
// specialist heuristics, tile-improvement spending or national slider optimizer.
func fillGovernorQueue(state *State, city *City, runtime *rules.RuntimeRules) error {
	if state.Terminal.Ended || city.Governor == nil || !city.Governor.Enabled || len(city.Queue) > 0 || runtime == nil || runtime.AI == nil {
		return nil
	}
	if player, ok := playerByIDValue(*state, city.OwnerID); !ok || player.Lifecycle != PlayerActive {
		return nil
	}
	if city.QueueSequence == math.MaxUint64 {
		return errors.New("production identity sequence exhausted")
	}
	priorities := append([]string{city.Governor.Priority}, governorPriorities...)
	seen := map[string]bool{}
	for _, kind := range []string{ProductionBuilding, ProductionWonder} {
		lists := runtime.AI.BuildingBuildLists
		prefix := "BUILDING_BUILD_LIST_"
		if kind == ProductionWonder {
			lists = runtime.AI.WonderBuildLists
			prefix = "WONDER_BUILD_LIST_"
		}
		for _, priority := range priorities {
			for _, list := range lists {
				if list.ID != prefix+strings.ToUpper(priority) {
					continue
				}
				for _, id := range list.Items {
					if seen[id] {
						continue
					}
					seen[id] = true
					item, err := canonicalProductionItem(runtime, id)
					if err != nil || item.Kind != kind {
						continue
					}
					if err := queueProduction(state, city.OwnerID, QueueProductionPayload{CityID: city.ID, Item: item}, runtime); err != nil {
						continue
					}
					selected := city.Queue[0]
					appendNotification(state, Notification{ID: "governor:" + selected.EntryID, PlayerID: city.OwnerID, Kind: "governor.production_selected", Turn: state.Phase.Turn, CityID: city.ID, TileID: city.TileID, ContentID: selected.ID})
					return nil
				}
			}
		}
	}
	return nil
}

func validateGovernorState(city City) error {
	if city.Governor == nil {
		return nil
	}
	if city.Governor.Priority == "" && !city.Governor.Enabled {
		return nil
	}
	if !containsString(governorPriorities, city.Governor.Priority) {
		return errors.New("invalid governor priority")
	}
	return nil
}
