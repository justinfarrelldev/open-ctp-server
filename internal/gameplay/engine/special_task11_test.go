package engine

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
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

func TestRetailPillageRemovesCompletedEnemyTileImprovements(t *testing.T) {
	state := task11State()
	state.Tiles[0].OwnerID = "player-b"
	state.Units[0].Effects = []Effect{{ID: "pillage", Kind: EffectCanPillage}}
	state.Overlays = []TileOverlay{{ID: "farm", TileID: "a", Kind: "improvement", OwnerID: "player-b", ImprovementID: "TILEIMP_FARMS", YieldFood: 1}, {ID: "pending", TileID: "a", Kind: "improvement", OwnerID: "player-b", ImprovementID: "TILEIMP_MINES", Remaining: 1}}

	next, events, err := task11Apply(t, state, CommandPillage, PillagePayload{UnitID: "agent"})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventPillaged {
		t.Fatalf("pillage events = %+v", events)
	}
	if len(next.Overlays) != 1 || next.Overlays[0].ID != "pending" {
		t.Fatalf("overlays after pillage = %+v", next.Overlays)
	}
	if next.Units[0].Movement != 3 {
		t.Fatalf("pillage did not spend movement: %+v", next.Units[0])
	}
}

func TestRetailInjoinCityBlocksProduction(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = []Effect{{ID: "injoin", Kind: EffectSpecialInjoin}}
	next, events, err := task11Apply(t, state, CommandInjoin, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if !hasEffect(next.Cities[0].Effects, EffectInjunction) || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventCityInjoined || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("injoin outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
	}

	next.Cities[0].Queue = []ProductionItem{{ID: "UNIT_TEST", Kind: ProductionUnit, Cost: 1}}
	next.Cities[0].Yields.Production = 3
	if err := completeProduction(&next, &next.Cities[0], &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_TEST", Domain: DomainLand, Movement: 1, Health: 1}}}); err != nil {
		t.Fatal(err)
	}
	if _, ok := unitByID(next, "unit-UNIT_TEST-target-city-7"); ok || len(next.Cities[0].Queue) != 1 {
		t.Fatalf("injunction did not block production: city=%+v units=%+v", next.Cities[0], next.Units)
	}
}

func TestRetailExpelMovesAdjacentEnemyUnit(t *testing.T) {
	state := task11State()
	state.Tiles = append(state.Tiles, Tile{ID: "c", X: 2, Y: 0, Terrain: TerrainPlains, OwnerID: "player-b", Open: true})
	state.Units = append(state.Units, Unit{ID: "intruder", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Movement: 1})
	state.Units[0].Effects = []Effect{{ID: "expel", Kind: EffectSpecialExpel}}
	next, events, err := task11Apply(t, state, CommandExpel, ExpelPayload{UnitID: "agent", TargetUnitID: "intruder"})
	if err != nil {
		t.Fatal(err)
	}
	target, ok := unitByID(next, "intruder")
	if !ok || target.TileID != "c" || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventUnitExpelled || strings.Contains(string(events[0].Payload), "intruder") {
		t.Fatalf("expel outcome target=%+v ok=%v unit=%+v events=%+v", target, ok, next.Units[0], events)
	}
}

func TestRetailAssassinateRulerDestabilizesTargetNation(t *testing.T) {
	state := task11State()
	state.Nations[1].GovernmentID = "GOVERNMENT_MONARCHY"
	state.Nations[1].ResearchProgress = 12
	state.Units[0].Effects = []Effect{{ID: "assassin", Kind: EffectSpecialAssassinateRuler}}
	next, events, err := task11Apply(t, state, CommandAssassinateRuler, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Nations[1].GovernmentID != "" || next.Nations[1].ResearchProgress != 0 || !hasEffect(next.Nations[1].Effects, EffectRulerAssassinated) || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventRulerAssassinated || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("assassination outcome nation=%+v unit=%+v events=%+v", next.Nations[1], next.Units[0], events)
	}
}

func TestRetailCreateParkCleansPollutedAdjacentTile(t *testing.T) {
	state := task11State()
	state.Tiles[1].Pollution = 7
	state.Units[0].Effects = []Effect{{ID: "park", Kind: EffectSpecialCreatePark}}
	next, events, err := task11Apply(t, state, CommandCreatePark, CreateParkPayload{UnitID: "agent", TileID: "b"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Tiles[1].Pollution != 0 || next.Units[0].Movement != 3 || len(next.Overlays) != 1 || next.Overlays[0].Kind != "park" || len(events) != 1 || events[0].Type != EventParkCreated {
		t.Fatalf("park outcome tile=%+v unit=%+v overlays=%+v events=%+v", next.Tiles[1], next.Units[0], next.Overlays, events)
	}

	_, _, err = task11Apply(t, next, CommandCreatePark, CreateParkPayload{UnitID: "agent", TileID: "b"})
	if err == nil {
		t.Fatal("duplicate park accepted")
	}
}

func TestRetailPlantNukeDamagesCityAndHonorsProtection(t *testing.T) {
	state := task11State()
	state.Cities[0].Population = 8
	state.Units[0].Effects = []Effect{{ID: "nuke", Kind: EffectSpecialPlantNuke}}
	next, events, err := task11Apply(t, state, CommandPlantNuke, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].Population != 4 || next.Tiles[1].Pollution == 0 || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventNukePlanted || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("nuke outcome city=%+v tile=%+v unit=%+v events=%+v", next.Cities[0], next.Tiles[1], next.Units[0], events)
	}

	state = task11State()
	state.Cities[0].Population = 8
	state.Cities[0].Effects = append(state.Cities[0].Effects, Effect{ID: "force", Kind: EffectForceField})
	state.Units[0].Effects = []Effect{{ID: "nuke", Kind: EffectSpecialPlantNuke}}
	next, events, err = task11Apply(t, state, CommandPlantNuke, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].Population != 8 || next.Tiles[1].Pollution != 0 || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventNukeFailed {
		t.Fatalf("protected nuke outcome city=%+v tile=%+v unit=%+v events=%+v", next.Cities[0], next.Tiles[1], next.Units[0], events)
	}
}

func TestRetailInvestigateCityStoresPrivateCityReport(t *testing.T) {
	state := task11State()
	state.Cities[0].Population = 6
	state.Cities[0].Buildings = []Building{{ID: "bank"}, {ID: "walls"}}
	state.Cities[0].Effects = []Effect{{ID: "defense", Kind: EffectCombatDefense, Value: 2}}
	state.Units[0].Effects = []Effect{{ID: "investigate", Kind: EffectSpecialInvestigateCity}}

	next, events, err := task11Apply(t, state, CommandInvestigateCity, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventCityInvestigated || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("investigate outcome unit=%+v events=%+v", next.Units[0], events)
	}
	if got := next.Nations[0].CityReports; len(got) != 1 || got[0].CityID != "target-city" || got[0].Population != 6 || len(got[0].Buildings) != 2 || len(got[0].Effects) != 1 {
		t.Fatalf("city reports = %+v", got)
	}

	next.Cities[0].Population = 7
	next.Units[0].Movement = 1
	next, _, err = task11Apply(t, next, CommandInvestigateCity, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if got := next.Nations[0].CityReports; len(got) != 1 || got[0].Population != 7 {
		t.Fatalf("city report was not updated in place: %+v", got)
	}
}

func TestRetailCreateFranchiseMarksCityOnce(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = []Effect{{ID: "franchise", Kind: EffectSpecialCreateFranchise}}
	next, events, err := task11Apply(t, state, CommandCreateFranchise, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Units[0].Movement != 3 || len(next.Cities[0].Effects) != 1 || next.Cities[0].Effects[0].Kind != EffectFranchise || len(events) != 1 || events[0].Type != EventFranchiseCreated || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("franchise outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
	}

	next, _, err = task11Apply(t, next, CommandCreateFranchise, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err == nil {
		t.Fatalf("duplicate franchise accepted: %+v", next.Cities[0].Effects)
	}
}

func TestRetailSlaveRaidsReducePopulationAndHonorProtection(t *testing.T) {
	for _, tc := range []struct {
		name   string
		kind   string
		effect string
	}{
		{name: "slave", kind: CommandSlaveRaid, effect: EffectSpecialSlaveRaid},
		{name: "settler", kind: CommandSettlerSlaveRaid, effect: EffectSpecialSettlerSlaveRaid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := task11State()
			state.Cities[0].Population = 3
			state.Units[0].Effects = []Effect{{ID: tc.name, Kind: tc.effect}}
			next, events, err := task11Apply(t, state, tc.kind, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
			if err != nil {
				t.Fatal(err)
			}
			if next.Cities[0].Population != 2 || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventSlaveRaidApplied || strings.Contains(string(events[0].Payload), "target-city") {
				t.Fatalf("slave raid outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
			}

			state = task11State()
			state.Cities[0].Population = 3
			state.Cities[0].Effects = append(state.Cities[0].Effects, Effect{ID: "anti-slave", Kind: EffectPreventSlavery, Value: 100})
			state.Units[0].Effects = []Effect{{ID: tc.name, Kind: tc.effect}}
			next, events, err = task11Apply(t, state, tc.kind, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
			if err != nil {
				t.Fatal(err)
			}
			if next.Cities[0].Population != 3 || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventSlaveRaidFailed {
				t.Fatalf("protected slave raid outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
			}
		})
	}
}

func TestRetailBioNanoAndPlagueSpecialsAffectCityPopulation(t *testing.T) {
	for _, tc := range []struct {
		name       string
		command    string
		effectKind string
		protection string
		eventType  string
	}{
		{name: "bio", command: CommandBioTerror, effectKind: EffectSpecialBioTerror, protection: EffectProtectFromBioAgents, eventType: EventBioTerrorApplied},
		{name: "nano", command: CommandNanoTerror, effectKind: EffectSpecialNanoTerror, protection: EffectProtectFromNanoVirus, eventType: EventNanoTerrorApplied},
		{name: "plague", command: CommandPlague, effectKind: EffectSpecialPlague, eventType: EventPlagueApplied},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := task11State()
			state.Cities[0].Population = 5
			state.Units[0].Effects = []Effect{{ID: tc.name, Kind: tc.effectKind}}
			next, events, err := task11Apply(t, state, tc.command, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
			if err != nil {
				t.Fatal(err)
			}
			if next.Cities[0].Population != 4 || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != tc.eventType || strings.Contains(string(events[0].Payload), "target-city") {
				t.Fatalf("terror outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
			}

			if tc.protection == "" {
				return
			}
			state = task11State()
			state.Cities[0].Population = 5
			state.Cities[0].Effects = append(state.Cities[0].Effects, Effect{ID: "protect", Kind: tc.protection, Value: 100})
			state.Units[0].Effects = []Effect{{ID: tc.name, Kind: tc.effectKind}}
			next, events, err = task11Apply(t, state, tc.command, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
			if err != nil {
				t.Fatal(err)
			}
			if next.Cities[0].Population != 5 || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventTerrorFailed {
				t.Fatalf("protected terror outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
			}
		})
	}
}

func TestRetailConvertCityHonorsProtectionAndDisclosure(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = append(state.Units[0].Effects, Effect{ID: "convert", Kind: EffectSpecialConvertCity})

	next, events, err := task11Apply(t, state, CommandConvertCity, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].OwnerID != "player-a" || next.Units[0].Movement != 3 {
		t.Fatalf("conversion outcome city=%+v unit=%+v", next.Cities[0], next.Units[0])
	}
	if len(events) != 1 || events[0].Type != EventCityConverted || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("unsafe conversion event=%+v", events)
	}

	state = task11State()
	state.Units[0].Effects = append(state.Units[0].Effects, Effect{ID: "convert", Kind: EffectSpecialConvertCity})
	state.Cities[0].Effects = append(state.Cities[0].Effects, Effect{ID: "walls", Kind: EffectPreventConversion, Value: 100})
	next, events, err = task11Apply(t, state, CommandConvertCity, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].OwnerID != "player-b" || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventCityConversionFailed {
		t.Fatalf("protected conversion outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
	}
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
