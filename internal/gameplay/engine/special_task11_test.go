package engine

import (
	"encoding/json"
	"strings"
	"testing"
)

func task11State() State {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "a", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true}, {ID: "b", X: 1, Y: 0, Terrain: TerrainForest, OwnerID: "player-b", Open: true}}
	state.Units = []Unit{{ID: "agent", OwnerID: "player-a", TileID: "a", Domain: DomainLand, Movement: 4, Effects: []Effect{{ID: "embassy", Kind: EffectSpecialEmbassy}, {ID: "soothsay", Kind: EffectSpecialSoothsaying}, {ID: "steal", Kind: EffectSpecialSteal, Value: 100}, {ID: "capture", Kind: EffectSpecialCapture, Value: 100}, {ID: "terraform", Kind: EffectSpecialTerraform, Value: 3}, {ID: "terraform-turns", Kind: EffectSpecialTerraformTurns, Value: 2}}}}
	state.Cities = []City{{ID: "target-city", OwnerID: "player-b", TileID: "b"}}
	state.Nations = []Nation{{ID: "nation-a", PublicWorks: 5}, {ID: "nation-b", Advances: []string{"zeta", "alpha"}}}
	return state
}

func task11Apply(t *testing.T, state State, kind string, payload any) (State, []Event, error) {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	next, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "task11-" + kind, ActorPlayerID: "player-a", Type: kind, Payload: body}, state.Rules, fixtureSources())
	return next, events, err
}

func TestTask11EmbassySoothsayingAndDisclosureVectors(t *testing.T) {
	state := task11State()
	next, events, err := task11Apply(t, state, CommandEstablishEmbassy, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if got := next.Nations[0].Embassies; len(got) != 1 || got[0] != "nation-b" {
		t.Fatalf("embassy=%v", got)
	}
	if events[0].Type != EventEmbassyEstablished || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("unsafe embassy event=%s", events[0].Payload)
	}
	next, events, err = task11Apply(t, state, CommandSoothsaying, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Cities[0].Effects) != 1 || next.Cities[0].Effects[0].Kind != EffectSoothsaying {
		t.Fatalf("city effects=%+v", next.Cities[0].Effects)
	}
	if events[0].Type != EventSoothsayingApplied || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("unsafe soothsaying event=%s", events[0].Payload)
	}
}

func TestTask11StealTechnologySortedRNGFailureCaptureAndDisclosure(t *testing.T) {
	state := task11State()
	next, events, err := task11Apply(t, state, CommandStealTechnology, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if got := next.Nations[0].Advances; len(got) != 1 || got[0] != "alpha" {
		t.Fatalf("advances=%v", got)
	}
	if nextState, _ := next.RNGState(RNGDomainSpecialAction); func() bool { before, _ := state.RNGState(RNGDomainSpecialAction); return nextState == before }() {
		t.Fatal("special RNG was not consumed")
	}
	if events[0].Type != EventTechnologyStolen || strings.Contains(string(events[0].Payload), "alpha") || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("unsafe steal event=%s", events[0].Payload)
	}

	state.Units[0].Effects[2].Value = 0
	next, events, err = task11Apply(t, state, CommandStealTechnology, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Units) != 0 || events[0].Type != EventTechnologyCapture {
		t.Fatalf("capture outcome units=%+v event=%s", next.Units, events[0].Type)
	}
}

func TestTask11SpecialFailureAndStateValidationVectors(t *testing.T) {
	state := task11State()
	state.Units[0].Effects[2].Value = 0
	state.Units[0].Effects[3].Value = 0
	next, events, err := task11Apply(t, state, CommandStealTechnology, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil || len(next.Units) != 1 || events[0].Type != EventTechnologyFailed {
		t.Fatalf("non-capture failure next=%+v event=%+v err=%v", next.Units, events, err)
	}
	state = task11State()
	state.Nations[0].PublicWorks = -1
	if err := ValidateState(state); err == nil {
		t.Fatal("negative public works accepted")
	}
	state = task11State()
	state.Overlays = []TileOverlay{{ID: "bad", TileID: "b", Kind: "terraform", OwnerID: "player-a", Remaining: -1, TargetTerrain: TerrainPlains}}
	if err := ValidateState(state); err == nil {
		t.Fatal("negative terraform duration accepted")
	}
	state = task11State()
	state.Nations[0].Embassies = []string{"nation-b", "nation-b"}
	if err := ValidateState(state); err == nil {
		t.Fatal("duplicate embassy knowledge accepted")
	}
}

func TestTask11TerraformQueueCompletionCancellationAndImmutableTiles(t *testing.T) {
	state := task11State()
	baseTerrain := state.Tiles[1].Terrain
	next, _, err := task11Apply(t, state, CommandTerraform, TerraformPayload{UnitID: "agent", TileID: "b", TargetTerrain: TerrainPlains})
	if err != nil {
		t.Fatal(err)
	}
	if next.Tiles[1].Terrain != baseTerrain || len(next.Overlays) != 1 || next.Overlays[0].Remaining != 2 || next.Nations[0].PublicWorks != 2 {
		t.Fatalf("invalid queued terraform=%+v", next)
	}
	if got := EffectiveTerrain(next, "b"); got != TerrainForest {
		t.Fatalf("unfinished terrain=%q", got)
	}
	queued, err := RecalculateRound(next)
	if err != nil {
		t.Fatal(err)
	}
	queued.Phase.Turn++
	complete, err := RecalculateRound(queued)
	if err != nil {
		t.Fatal(err)
	}
	effective := EffectiveTerrain(complete, "b")
	if complete.Tiles[1].Terrain != baseTerrain || effective != TerrainPlains {
		t.Fatalf("completion terrain base=%q effective=%q", complete.Tiles[1].Terrain, effective)
	}
	if _, _, err := task11Apply(t, next, CommandCancelTerraform, CancelTerraformPayload{TileID: "b"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := task11Apply(t, state, CommandTerraform, TerraformPayload{UnitID: "agent", TileID: "a", TargetTerrain: TerrainOcean}); err == nil {
		t.Fatal("illegal transition accepted")
	}
}

func TestTask11RejectsCapabilityRangeCostAndHiddenTargetWithoutLeaks(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = nil
	_, _, err := task11Apply(t, state, CommandEstablishEmbassy, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err == nil || strings.Contains(err.Error(), "target-city") {
		t.Fatalf("leaking capability error=%v", err)
	}
	state = task11State()
	state.Nations[0].PublicWorks = 2
	_, _, err = task11Apply(t, state, CommandTerraform, TerraformPayload{UnitID: "agent", TileID: "b", TargetTerrain: TerrainPlains})
	if err == nil {
		t.Fatalf("cost rejection missing: %v", err)
	}
	state = task11State()
	_, _, err = task11Apply(t, state, CommandStealTechnology, SpecialActionPayload{UnitID: "agent", CityID: "hidden-city"})
	if err == nil || strings.Contains(err.Error(), "hidden-city") {
		t.Fatalf("leaking hidden error=%v", err)
	}
}
