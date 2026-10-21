package engine

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestClientRepeatedUnitsHaveIndependentStableQueueIdentity(t *testing.T) {
	state := task10State()
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 2}}
	item := ProductionItem{ID: "warrior", Kind: ProductionUnit, Cost: 100}
	state = task10Apply(t, state, "player-a", CommandQueueProduction, QueueProductionPayload{CityID: "home", Item: item})
	state = task10Apply(t, state, "player-a", CommandQueueProduction, QueueProductionPayload{CityID: "home", Item: item})
	queue := state.Cities[0].Queue
	if len(queue) != 2 || queue[0].EntryID == "" || queue[0].EntryID == queue[1].EntryID {
		t.Fatalf("missing independent identity: %+v", queue)
	}
	before, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := cancelProduction(&state, "player-a", CancelProductionPayload{CityID: "home", ItemID: "warrior"}); err == nil {
		t.Fatal("ambiguous legacy cancellation accepted")
	}
	after, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("ambiguous cancellation changed state")
	}
	state = task10Apply(t, state, "player-a", CommandCancelProduction, CancelProductionPayload{CityID: "home", EntryID: queue[1].EntryID})
	if len(state.Cities[0].Queue) != 1 || state.Cities[0].Queue[0].EntryID != queue[0].EntryID {
		t.Fatal("wrong duplicate removed")
	}
}

func TestClientAtomicQueueReorderClearAndInvalidRollback(t *testing.T) {
	state := task10State()
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 2, StoredProduction: 40, Queue: []ProductionItem{{ID: "head", Kind: ProductionUnit, Cost: 100}, {ID: "tail", Kind: ProductionUnit, Cost: 100}, {ID: "middle", Kind: ProductionUnit, Cost: 100}}}}
	var err error
	state, err = WithQueueIdentities(state)
	if err != nil {
		t.Fatal(err)
	}
	queue := state.Cities[0].Queue
	edit := EditProductionPayload{CityID: "home", Entries: []QueueSelection{{EntryID: queue[0].EntryID}, {EntryID: queue[2].EntryID}, {EntryID: queue[1].EntryID}}}
	state = task10Apply(t, state, "player-a", CommandEditProduction, edit)
	if state.Cities[0].StoredProduction != 40 || state.Cities[0].Queue[1].EntryID != queue[2].EntryID {
		t.Fatalf("future reorder lost progress/identity: %+v", state.Cities[0])
	}
	before, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(EditProductionPayload{CityID: "home", Entries: []QueueSelection{{EntryID: queue[0].EntryID}, {EntryID: "missing"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "bad-edit", ActorPlayerID: "player-a", Type: CommandEditProduction, Payload: body}, state.Rules, fixtureSources()); err == nil {
		t.Fatal("invalid edit accepted")
	}
	after, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("invalid edit changed authority")
	}
	state = task10Apply(t, state, "player-a", CommandEditProduction, EditProductionPayload{CityID: "home", Entries: []QueueSelection{{EntryID: queue[1].EntryID}, {EntryID: queue[0].EntryID}}})
	if state.Cities[0].StoredProduction != 0 {
		t.Fatal("head change failed explicit no-refund policy")
	}
	state = task10Apply(t, state, "player-a", CommandEditProduction, EditProductionPayload{CityID: "home", Entries: []QueueSelection{}})
	if len(state.Cities[0].Queue) != 0 {
		t.Fatal("atomic clear failed")
	}
}

func TestClientPinnedProductionCannotRedefineRules(t *testing.T) {
	state := task10State()
	ruleset, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: ruleset.CanonicalJSON}
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 2}}
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "home", Item: ProductionItem{ID: "forged", Kind: ProductionUnit, Cost: 1}}, &runtime); err == nil {
		t.Fatal("unknown client-defined production accepted")
	}
	unit := runtime.Starter.TypeID
	if unit == "" {
		unit = runtime.Units[0].ID
	}
	state = task10Apply(t, state, "player-a", CommandEditProduction, EditProductionPayload{CityID: "home", Entries: []QueueSelection{{ItemID: unit}, {ItemID: unit}}})
	if len(state.Cities[0].Queue) != 2 || state.Cities[0].Queue[0].Cost < 1 {
		t.Fatal("canonical insertion failed")
	}
	for _, item := range state.Cities[0].Queue {
		if item.ID != unit || item.EntryID == "" {
			t.Fatalf("canonical item wrong: %+v", item)
		}
	}
}

func TestClientQueueEditRejectsMalformedAndCrossCitySelectionsAtomically(t *testing.T) {
	state := task10State()
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", StoredProduction: 40, Queue: []ProductionItem{{ID: "warrior", Kind: ProductionUnit, Cost: 100, EntryID: "home:queue:10"}}}}
	for _, payload := range []string{
		`{"city_id":"home"}`,
		`{"city_id":"home","entries":null}`,
		`{"city_id":"home","entries":[{"entry_id":"foreign:queue:1"}]}`,
		`{"city_id":"home","entries":[{"entry_id":"home:queue:10"},{"entry_id":"home:queue:10"}]}`,
		`{"city_id":"home","entries":[{"entry_id":"home:queue:10","item_id":"forged"}]}`,
		`{"city_id":"home","entries":[{"item_id":"forged","cost":1}]}`,
	} {
		t.Run(payload, func(t *testing.T) {
			before, err := CanonicalStateJSON(state)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "invalid", ActorPlayerID: "player-a", Type: CommandEditProduction, Payload: []byte(payload)}, state.Rules, fixtureSources()); err == nil {
				t.Fatal("invalid edit accepted")
			}
			after, err := CanonicalStateJSON(state)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatal("invalid edit changed state")
			}
		})
	}
	identified, err := WithQueueIdentities(state)
	if err != nil {
		t.Fatal(err)
	}
	if identified.Cities[0].QueueSequence != 10 {
		t.Fatal("legacy counter did not catch up with existing identities")
	}
}

func TestClientQueueCompletionPreservesRemainingDuplicateIdentity(t *testing.T) {
	state := task10State()
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 2, Queue: []ProductionItem{{ID: "warrior", Kind: ProductionUnit, Cost: 1}, {ID: "warrior", Kind: ProductionUnit, Cost: 1}}}}
	var err error
	state, err = WithQueueIdentities(state)
	if err != nil {
		t.Fatal(err)
	}
	second := state.Cities[0].Queue[1].EntryID
	if err := completeProduction(&state, &state.Cities[0], nil); err != nil {
		t.Fatal(err)
	}
	if len(state.Cities[0].Queue) != 1 || state.Cities[0].Queue[0].EntryID != second {
		t.Fatalf("completion removed wrong duplicate: %+v", state.Cities[0].Queue)
	}
	raw, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Cities[0].Queue[0].EntryID != second {
		t.Fatal("restart changed entry identity")
	}
}

func TestClientCancelFutureProductionPreservesCurrentProgress(t *testing.T) {
	state := task10State()
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", StoredProduction: 40, Queue: []ProductionItem{{ID: "hoplite", Kind: ProductionUnit, Cost: 100}, {ID: "granary", Kind: ProductionBuilding, Cost: 50}}}}
	if err := cancelProduction(&state, "player-a", CancelProductionPayload{CityID: "home", ItemID: "granary"}); err != nil {
		t.Fatal(err)
	}
	if state.Cities[0].StoredProduction != 40 || len(state.Cities[0].Queue) != 1 || state.Cities[0].Queue[0].ID != "hoplite" {
		t.Fatalf("tail removal destroyed current production: %+v", state.Cities[0])
	}
}
