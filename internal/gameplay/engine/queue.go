package engine

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// MaxProductionQueueEntries bounds atomic edits and legacy appends alike.
const MaxProductionQueueEntries = 256

// QueueSelection retains an existing entry OR inserts a canonical content ID.
// It cannot supply costs, effects, prerequisites or a new entry's identity.
type QueueSelection struct {
	EntryID string `json:"entry_id,omitempty"`
	ItemID  string `json:"item_id,omitempty"`
}

// EditProductionPayload replaces the ordered list atomically. Explicit [] clears;
// omitted/null entries are invalid, preventing accidental destructive requests.
type EditProductionPayload struct {
	CityID  string           `json:"city_id"`
	Entries []QueueSelection `json:"entries"`
}

// WithQueueIdentities upgrades legacy rows without mutating the supplied state.
// Read projections derive the same IDs that the next accepted command persists.
// Monotonic per-city sequences prevent reuse after removal/completion/reconnect.
func WithQueueIdentities(state State) (State, error) {
	state.Cities = append([]City(nil), state.Cities...)
	for index := range state.Cities {
		city, err := identifiedQueue(state.Cities[index])
		if err != nil {
			return State{}, err
		}
		state.Cities[index] = city
	}
	return state, nil
}

func identifiedQueue(city City) (City, error) {
	city.Queue = append([]ProductionItem(nil), city.Queue...)
	seen := make(map[string]bool, len(city.Queue))
	for _, item := range city.Queue {
		if item.EntryID == "" {
			continue
		}
		if seen[item.EntryID] {
			return City{}, errors.New("duplicate production entry identity")
		}
		seen[item.EntryID] = true
		prefix := city.ID + ":queue:"
		if strings.HasPrefix(item.EntryID, prefix) {
			if sequence, err := strconv.ParseUint(strings.TrimPrefix(item.EntryID, prefix), 10, 64); err == nil && sequence > city.QueueSequence {
				city.QueueSequence = sequence
			}
		}
	}
	for index := range city.Queue {
		city.Queue[index].Effects = append([]Effect(nil), city.Queue[index].Effects...)
		if city.Queue[index].EntryID != "" {
			continue
		}
		for {
			if city.QueueSequence == math.MaxUint64 {
				return City{}, errors.New("production identity sequence exhausted")
			}
			city.QueueSequence++
			id := city.ID + ":queue:" + strconv.FormatUint(city.QueueSequence, 10)
			if seen[id] {
				continue
			}
			city.Queue[index].EntryID = id
			seen[id] = true
			break
		}
	}
	return city, nil
}

// canonicalProductionItem is the only source of executable production metadata
// in pinned games. Even legacy queue-production requests cannot redefine rules.
func canonicalProductionItem(runtime *rules.RuntimeRules, id string) (ProductionItem, error) {
	if runtime == nil {
		return ProductionItem{}, productionFailure("content_unavailable", "canonical production rules are unavailable")
	}
	item := ProductionItem{ID: id}
	if production, ok := runtime.ProductionDefinition(id); ok {
		item.Kind = production.Kind
		item.Cost = production.Cost
		item.Upkeep = production.Upkeep
		item.Effects = runtimeProductionEffects(production)
		if len(production.Prerequisites) > 0 {
			item.Prerequisite = production.Prerequisites[0]
		}
		if len(production.ObsoleteBy) > 0 {
			item.ObsoletedBy = production.ObsoleteBy[0]
		}
	}
	unit, hasUnit := runtime.Unit(id)
	if item.Kind == ProductionUnit && !hasUnit {
		return ProductionItem{}, productionFailure("content_unavailable", "production unit definition is unavailable")
	}
	if hasUnit {
		item.Kind = ProductionUnit
		if unit.ProductionCost > 0 {
			item.Cost = unit.ProductionCost
		}
		item.Upkeep = unit.Upkeep
		if len(unit.Prerequisites) > 0 {
			item.Prerequisite = unit.Prerequisites[0]
		}
		if len(unit.ObsoleteBy) > 0 {
			item.ObsoletedBy = unit.ObsoleteBy[0]
		}
	}
	// Synthetic runtimes historically omit catalog costs for known units;
	// the server-owned supported default is one, matching legal-action output.
	// Unknown IDs never receive this default.
	if item.Kind != "" && item.Cost < 1 {
		item.Cost = 1
	}
	if err := validateProductionItem(item); err != nil {
		return ProductionItem{}, productionFailure("content_unavailable", "production content is unavailable")
	}
	return item, nil
}

func editProduction(state *State, actor string, payload EditProductionPayload, runtime *rules.RuntimeRules) error {
	city, ok := cityByID(*state, payload.CityID)
	if !ok || city.OwnerID != actor {
		return productionFailure("queue_identity", "production city identity is invalid")
	}
	if payload.Entries == nil {
		return productionFailure("queue_shape", "production entries must be an explicit array")
	}
	if len(payload.Entries) > MaxProductionQueueEntries {
		return productionFailure("queue_capacity", "production queue is full")
	}
	identified, err := identifiedQueue(*city)
	if err != nil {
		return err
	}
	existing := make(map[string]ProductionItem, len(identified.Queue))
	for _, item := range identified.Queue {
		existing[item.EntryID] = item
	}
	candidate := *state
	candidate.Cities = append([]City(nil), state.Cities...)
	target, _ := cityByID(candidate, city.ID)
	*target = identified
	target.Queue = nil
	used := make(map[string]bool, len(payload.Entries))
	for _, selection := range payload.Entries {
		if (selection.EntryID == "") == (selection.ItemID == "") {
			return productionFailure("queue_shape", "queue selection requires exactly one identity")
		}
		var item ProductionItem
		if selection.EntryID != "" {
			var found bool
			item, found = existing[selection.EntryID]
			if !found || used[selection.EntryID] {
				return productionFailure("queue_identity", "queue entry is missing or duplicated")
			}
			used[selection.EntryID] = true
		} else {
			item, err = canonicalProductionItem(runtime, selection.ItemID)
			if err != nil {
				return err
			}
		}
		retainedID := item.EntryID
		item.EntryID = ""
		sequence := target.QueueSequence
		if err := queueProduction(&candidate, actor, QueueProductionPayload{CityID: city.ID, Item: item}, runtime); err != nil {
			return err
		}
		// Validation and rule resolution are shared with ordinary append. Retained
		// rows keep their identities; only genuinely new rows consume a sequence.
		if retainedID != "" {
			target.Queue[len(target.Queue)-1].EntryID = retainedID
			target.QueueSequence = sequence
		}
	}
	oldHead, newHead := "", ""
	if len(identified.Queue) > 0 {
		oldHead = identified.Queue[0].EntryID
	}
	if len(target.Queue) > 0 {
		newHead = target.Queue[0].EntryID
	}
	if oldHead != "" && oldHead != newHead {
		target.StoredProduction = 0
	}
	// No candidate state escapes until every entry is validated successfully.
	*city = *target
	return nil
}
