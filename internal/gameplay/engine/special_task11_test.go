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

func TestRetailDefendAgainstSpiesStopsAndKillsCitySpy(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = append(state.Units[0].Effects, Effect{ID: "investigate", Kind: EffectSpecialInvestigateCity, Value: 100})
	state.Units = append(state.Units, Unit{ID: "counterspy", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Movement: 1, Effects: []Effect{{ID: "counterspy", Kind: EffectDefendAgainstSpies, Value: 25}}})

	next, events, err := task11Apply(t, state, CommandInvestigateCity, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatalf("investigate with counterspy: %v", err)
	}
	if _, ok := unitByID(next, "agent"); ok {
		t.Fatalf("acting spy survived spy defense: %#v", next.Units)
	}
	if _, ok := unitByID(next, "counterspy"); !ok {
		t.Fatalf("counterspy should survive spy defense: %#v", next.Units)
	}
	if len(events) != 1 || events[0].Type != EventCityInvestigationFailed || !strings.Contains(string(events[0].Payload), "stopped_by_spies") {
		t.Fatalf("spy defense event = %#v", events)
	}
	if hasEffect(next.Cities[0].Effects, EffectWatchful) {
		t.Fatalf("stopped spy action should not also mark city watchful: %#v", next.Cities[0].Effects)
	}
}

func TestRetailWatchfulCityHalvesSpecialSuccessChanceAndExpires(t *testing.T) {
	city := City{ID: "target", Effects: []Effect{{ID: "watchful", Kind: EffectWatchful, Value: 2}}}
	unitEffects := []Effect{{ID: "franchise", Kind: EffectSpecialCreateFranchise, Value: 75}}
	if got := citySpecialActionChance(city, unitEffects, EffectSpecialCreateFranchise); got != 37 {
		t.Fatalf("watchful chance = %d, want 37", got)
	}
	advanceTemporaryCityEffects(&city)
	if got := effectValue(city.Effects, EffectWatchful); got != 1 {
		t.Fatalf("watchful turns after first advance = %d, want 1", got)
	}
	advanceTemporaryCityEffects(&city)
	if hasEffect(city.Effects, EffectWatchful) {
		t.Fatalf("watchful effect did not expire: %+v", city.Effects)
	}
}

func TestRetailPlantNukeEscapeChanceTeleportsToNearestOwnCity(t *testing.T) {
	state := task11State()
	state.Tiles = append(state.Tiles, Tile{ID: "home-tile", X: 0, Y: 1, Terrain: TerrainPlains, OwnerID: "player-a", Open: true})
	state.Cities = append(state.Cities, City{ID: "home", OwnerID: "player-a", TileID: "home-tile"})
	state.Cities[0].Population = 8
	state.Units[0].Effects = []Effect{{ID: "plant", Kind: EffectSpecialPlantNuke, Value: 100}, {ID: "escape", Kind: EffectSpecialPlantNukeEscapeChance, Value: 100}}
	next, events, err := task11Apply(t, state, CommandPlantNuke, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	unit, ok := unitByID(next, "agent")
	if !ok || unit.TileID != "home-tile" || len(events) != 1 || events[0].Type != EventNukePlanted {
		t.Fatalf("plant nuke escape failed: unit=%+v ok=%v events=%+v", unit, ok, events)
	}
}

func TestRetailVeteranSpecialActionUsesEliteChance(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = []Effect{{ID: "investigate", Kind: EffectSpecialInvestigateCity, Value: 1}, {ID: "elite-investigate", Kind: EffectSpecialInvestigateCity + "_elite_chance", Value: 100}}
	failed, failedEvents, err := task11Apply(t, state, CommandInvestigateCity, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if len(failedEvents) != 1 || failedEvents[0].Type != EventCityInvestigationFailed || len(failed.Nations[0].CityReports) != 0 {
		t.Fatalf("non-veteran investigation used elite chance: reports=%+v events=%+v", failed.Nations[0].CityReports, failedEvents)
	}

	state.Units[0].Veteran = true
	succeeded, events, err := task11Apply(t, state, CommandInvestigateCity, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventCityInvestigated || len(succeeded.Nations[0].CityReports) != 1 {
		t.Fatalf("veteran investigation did not use elite chance: reports=%+v events=%+v", succeeded.Nations[0].CityReports, events)
	}
}

func TestRetailSpecialActionAttemptsSetCityWatchful(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = []Effect{{ID: "franchise", Kind: EffectSpecialCreateFranchise, Value: 10}}
	next, events, err := task11Apply(t, state, CommandCreateFranchise, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if !hasEffect(next.Cities[0].Effects, EffectWatchful) || effectValue(next.Cities[0].Effects, EffectWatchful) != 7 || len(events) != 1 || events[0].Type != EventFranchiseFailed {
		t.Fatalf("watchful franchise attempt city=%+v events=%+v", next.Cities[0], events)
	}
}

func TestRetailSpecialDeathChanceRemovesActingUnitAfterResolvedActions(t *testing.T) {
	for _, tc := range []struct {
		name        string
		command     string
		actionKind  string
		deathKind   string
		setup       func(*State)
		wantEvent   string
		assertState func(*testing.T, State)
	}{
		{name: "investigate", command: CommandInvestigateCity, actionKind: EffectSpecialInvestigateCity, deathKind: EffectSpecialInvestigateCityDeathChance, wantEvent: EventCityInvestigated, assertState: func(t *testing.T, state State) {
			if len(state.Nations[0].CityReports) != 1 {
				t.Fatalf("investigation report missing: %+v", state.Nations[0])
			}
		}},
		{name: "franchise", command: CommandCreateFranchise, actionKind: EffectSpecialCreateFranchise, deathKind: EffectSpecialCreateFranchiseDeathChance, wantEvent: EventFranchiseCreated, assertState: func(t *testing.T, state State) {
			if !hasCityFranchise(state.Cities[0], "player-a") {
				t.Fatalf("franchise marker missing: %+v", state.Cities[0])
			}
		}},
		{name: "plant-nuke", command: CommandPlantNuke, actionKind: EffectSpecialPlantNuke, deathKind: EffectSpecialPlantNukeDeathChance, wantEvent: EventNukePlanted, setup: func(state *State) { state.Cities[0].Population = 8 }, assertState: func(t *testing.T, state State) {
			if state.Cities[0].Population != 4 || state.Tiles[1].Pollution == 0 {
				t.Fatalf("nuke effects missing: city=%+v tile=%+v", state.Cities[0], state.Tiles[1])
			}
		}},
		{name: "assassinate", command: CommandAssassinateRuler, actionKind: EffectSpecialAssassinateRuler, deathKind: EffectSpecialAssassinateRulerDeathChance, wantEvent: EventRulerAssassinated, setup: func(state *State) { state.Nations[1].GovernmentID = "GOVERNMENT_MONARCHY" }, assertState: func(t *testing.T, state State) {
			if state.Nations[1].GovernmentID != "" || !hasEffect(state.Nations[1].Effects, EffectRulerAssassinated) {
				t.Fatalf("assassination effects missing: %+v", state.Nations[1])
			}
		}},
		{name: "underground", command: CommandUndergroundRailway, actionKind: EffectSpecialUndergroundRailway, deathKind: EffectSpecialUndergroundRailwayDeathChance, wantEvent: EventUndergroundRailway, setup: func(state *State) { state.Cities[0].Population = 4 }, assertState: func(t *testing.T, state State) {
			if state.Cities[0].Population != 3 || !hasEffect(state.Nations[0].Effects, EffectFreedPopulation) {
				t.Fatalf("underground railway effects missing: city=%+v nation=%+v", state.Cities[0], state.Nations[0])
			}
		}},
		{name: "slave", command: CommandSlaveRaid, actionKind: EffectSpecialSlaveRaid, deathKind: EffectSpecialSlaveRaidDeathChance, wantEvent: EventSlaveRaidApplied, setup: func(state *State) { state.Cities[0].Population = 4 }, assertState: func(t *testing.T, state State) {
			if state.Cities[0].Population != 3 {
				t.Fatalf("slave raid effects missing: %+v", state.Cities[0])
			}
		}},
		{name: "terror", command: CommandBioTerror, actionKind: EffectSpecialBioTerror, deathKind: EffectSpecialBioTerrorDeathChance, wantEvent: EventBioTerrorApplied, setup: func(state *State) { state.Cities[0].Population = 4 }, assertState: func(t *testing.T, state State) {
			if state.Cities[0].Population != 3 {
				t.Fatalf("terror effects missing: %+v", state.Cities[0])
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := task11State()
			if tc.setup != nil {
				tc.setup(&state)
			}
			state.Units[0].Effects = []Effect{{ID: "action", Kind: tc.actionKind}, {ID: "death", Kind: tc.deathKind, Value: 100}}
			next, events, err := task11Apply(t, state, tc.command, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := unitByID(next, "agent"); ok || len(events) != 1 || events[0].Type != tc.wantEvent {
				t.Fatalf("death chance outcome unit_present=%v events=%+v", ok, events)
			}
			tc.assertState(t, next)
		})
	}
}

func TestRetailSpecialProtectionsUseActiveCityBuildings(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = []Effect{{ID: "convert", Kind: EffectSpecialConvertCity}}
	state.Cities[0].Buildings = []Building{{ID: "IMPROVE_CITY_WALLS", Effects: []Effect{{ID: "prevent-conversion", Kind: EffectPreventConversion, Value: 100}}}}
	converted, events, err := task11Apply(t, state, CommandConvertCity, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if converted.Cities[0].OwnerID != "player-b" || len(events) != 1 || events[0].Type != EventCityConversionFailed {
		t.Fatalf("building conversion protection failed: city=%+v events=%+v", converted.Cities[0], events)
	}

	state = task11State()
	state.Units[0].Effects = []Effect{{ID: "nuke", Kind: EffectSpecialPlantNuke}}
	state.Cities[0].Population = 8
	state.Cities[0].Buildings = []Building{{ID: "IMPROVE_FORCE_FIELD", Effects: []Effect{{ID: "force-field", Kind: EffectForceField}}}}
	nuked, events, err := task11Apply(t, state, CommandPlantNuke, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if nuked.Cities[0].Population != 8 || len(events) != 1 || events[0].Type != EventNukeFailed {
		t.Fatalf("building nuke protection failed: city=%+v events=%+v", nuked.Cities[0], events)
	}
}

func TestRetailEmbassyWonderEffectsBlockRedundantEmbassyActions(t *testing.T) {
	state := task11State()
	state.Wonders = []Wonder{{ID: "WONDER_CLOSE", OwnerID: "player-b", CityID: "target-city", Effects: []Effect{{ID: "close", Kind: EffectCloseEmbassies}}}}
	if _, _, err := task11Apply(t, state, CommandEstablishEmbassy, SpecialActionPayload{UnitID: "agent", CityID: "target-city"}); err == nil || !strings.Contains(err.Error(), "special action") {
		t.Fatalf("close-embassies target allowed embassy err=%v", err)
	}

	state = task11State()
	state.Cities = append(state.Cities, City{ID: "home-city", OwnerID: "player-a", TileID: "a"})
	state.Wonders = []Wonder{{ID: "WONDER_GLOBAL_SURVEILLANCE_CENTER", OwnerID: "player-a", CityID: "home-city", Effects: []Effect{{ID: "everywhere", Kind: EffectEmbassiesEverywhere}}}}
	if _, _, err := task11Apply(t, state, CommandEstablishEmbassy, SpecialActionPayload{UnitID: "agent", CityID: "target-city"}); err == nil || !strings.Contains(err.Error(), "special action") {
		t.Fatalf("embassies-everywhere owner allowed redundant embassy err=%v", err)
	}
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

func TestRetailSueUnitRemovesCanBeSuedTargetsOnTile(t *testing.T) {
	state := task11State()
	state.Units = append(state.Units,
		Unit{ID: "branch-a", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Movement: 1, Effects: []Effect{{ID: "can-be-sued", Kind: EffectCanBeSued}}},
		Unit{ID: "branch-b", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Movement: 1, Effects: []Effect{{ID: "can-be-sued", Kind: EffectCanBeSued}}},
		Unit{ID: "guard", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Movement: 1},
	)
	state.Units[0].Effects = []Effect{{ID: "sue", Kind: EffectSpecialSue}}
	next, events, err := task11Apply(t, state, CommandSue, SpecialActionPayload{UnitID: "agent", TargetUnitID: "branch-a"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := unitByID(next, "branch-a"); ok {
		t.Fatalf("primary sued target survived: %+v", next.Units)
	}
	if _, ok := unitByID(next, "branch-b"); ok {
		t.Fatalf("secondary sued target on tile survived: %+v", next.Units)
	}
	if _, ok := unitByID(next, "guard"); !ok || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventUnitsSued || strings.Contains(string(events[0].Payload), "branch-a") {
		t.Fatalf("sue unit outcome units=%+v events=%+v", next.Units, events)
	}
}

func TestRetailSueUnitRequiresCanBeSuedTarget(t *testing.T) {
	state := task11State()
	state.Units = append(state.Units, Unit{ID: "guard", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Movement: 1})
	state.Units[0].Effects = []Effect{{ID: "sue", Kind: EffectSpecialSue}}
	_, _, err := task11Apply(t, state, CommandSue, SpecialActionPayload{UnitID: "agent", TargetUnitID: "guard"})
	if err == nil || !strings.Contains(err.Error(), "sue target") {
		t.Fatalf("sue without can-be-sued marker err=%v", err)
	}
}

func TestRetailSueCityAppliesLawsuit(t *testing.T) {
	state := task11State()
	state.Cities[0].StoredProduction = 7
	state.Units[0].Effects = []Effect{{ID: "sue", Kind: EffectSpecialSue}}
	next, events, err := task11Apply(t, state, CommandSue, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].StoredProduction != 0 || !hasEffect(next.Cities[0].Effects, EffectLawsuit) || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventCitySued {
		t.Fatalf("sue outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
	}
}

func TestRetailSueFranchiseRemovesOwnCityExternalFranchise(t *testing.T) {
	state := task11State()
	state.Cities[0].OwnerID = "player-a"
	state.Cities[0].Effects = []Effect{{ID: "franchise:player-b", Kind: EffectFranchise}}
	state.Units[0].Effects = []Effect{{ID: "sue", Kind: EffectSpecialSue}}
	next, events, err := task11Apply(t, state, CommandSue, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if hasEffect(next.Cities[0].Effects, EffectFranchise) || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventCitySued {
		t.Fatalf("sue franchise outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
	}
}

func TestRetailUndergroundRailwayFreesPopulation(t *testing.T) {
	state := task11State()
	state.Cities[0].Population = 4
	state.Units[0].Effects = []Effect{{ID: "railway", Kind: EffectSpecialUndergroundRailway}}
	next, events, err := task11Apply(t, state, CommandUndergroundRailway, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].Population != 3 || !hasEffect(next.Nations[0].Effects, EffectFreedPopulation) || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventUndergroundRailway {
		t.Fatalf("railway outcome city=%+v nation=%+v unit=%+v events=%+v", next.Cities[0], next.Nations[0], next.Units[0], events)
	}

	state = task11State()
	state.Cities[0].Population = 4
	state.Units[0].Effects = []Effect{{ID: "railway", Kind: EffectSpecialUndergroundRailway, Value: 10}}
	next, events, err = task11Apply(t, state, CommandUndergroundRailway, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].Population != 4 || hasEffect(next.Nations[0].Effects, EffectFreedPopulation) || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventUndergroundRailwayFailed {
		t.Fatalf("chance-failed railway outcome city=%+v nation=%+v unit=%+v events=%+v", next.Cities[0], next.Nations[0], next.Units[0], events)
	}
}

func TestRetailReformCityRemovesInjunctionAndCrime(t *testing.T) {
	state := task11State()
	state.Cities[0].CrimePercent = 35
	state.Cities[0].Happiness = 1
	state.Cities[0].Effects = []Effect{{ID: "injunction", Kind: EffectInjunction}}
	state.Units[0].Effects = []Effect{{ID: "reform", Kind: EffectSpecialReform}}
	next, events, err := task11Apply(t, state, CommandReform, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if hasEffect(next.Cities[0].Effects, EffectInjunction) || next.Cities[0].CrimePercent != 0 || next.Cities[0].Happiness != 2 || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventCityReformed {
		t.Fatalf("reform outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
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
	state.Units = append(state.Units, Unit{ID: "intruder", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Movement: 1, Effects: []Effect{{ID: "can-be-expelled", Kind: EffectCanBeExpelled}}})
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

func TestRetailExpelRequiresTargetCanBeExpelled(t *testing.T) {
	state := task11State()
	state.Tiles = append(state.Tiles, Tile{ID: "c", X: 2, Y: 0, Terrain: TerrainPlains, OwnerID: "player-b", Open: true})
	state.Units = append(state.Units, Unit{ID: "intruder", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Movement: 1})
	state.Units[0].Effects = []Effect{{ID: "expel", Kind: EffectSpecialExpel}}
	_, _, err := task11Apply(t, state, CommandExpel, ExpelPayload{UnitID: "agent", TargetUnitID: "intruder"})
	if err == nil || !strings.Contains(err.Error(), "expel target") {
		t.Fatalf("expel without can-be-expelled marker err=%v", err)
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

	state = task11State()
	state.Nations[1].GovernmentID = "GOVERNMENT_MONARCHY"
	state.Nations[1].ResearchProgress = 12
	state.Units[0].Effects = []Effect{{ID: "assassin", Kind: EffectSpecialAssassinateRuler, Value: 10}}
	next, events, err = task11Apply(t, state, CommandAssassinateRuler, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Nations[1].GovernmentID != "GOVERNMENT_MONARCHY" || next.Nations[1].ResearchProgress != 12 || hasEffect(next.Nations[1].Effects, EffectRulerAssassinated) || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventRulerAssassinationFailed {
		t.Fatalf("chance-failed assassination outcome nation=%+v unit=%+v events=%+v", next.Nations[1], next.Units[0], events)
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

func TestRetailPlantNukeRemovesUnitsInBlastRadiusAndRecordsDeathEffect(t *testing.T) {
	state := task11State()
	state.Tiles = append(state.Tiles, Tile{ID: "c", X: 1, Y: 1, Terrain: TerrainPlains, OwnerID: "player-b", Open: true})
	state.Units[0].Effects = []Effect{{ID: "nuke", Kind: EffectSpecialPlantNuke}}
	state.Units = append(state.Units,
		Unit{ID: "guard", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "death-happy", Kind: EffectDeathEffectsHappy}}},
		Unit{ID: "reserve", OwnerID: "player-b", TileID: "c", Domain: DomainLand, Health: 10, MaxHealth: 10},
	)
	next, events, err := task11Apply(t, state, CommandPlantNuke, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := unitByID(next, "guard"); ok {
		t.Fatalf("nuked guard survived: %#v", next.Units)
	}
	if _, ok := unitByID(next, "reserve"); ok {
		t.Fatalf("nuked reserve survived: %#v", next.Units)
	}
	if _, ok := unitByID(next, "agent"); !ok {
		t.Fatalf("acting plant-nuke unit should use existing escape/death mechanics, not blast removal: %#v", next.Units)
	}
	if len(events) != 1 || events[0].Type != EventNukePlanted || !strings.Contains(string(events[0].Payload), "destroyed_units") {
		t.Fatalf("nuke event = %#v", events)
	}
	if next.Nations[1].PeaceMovement == nil || next.Nations[1].PeaceMovement.UnitsLostHome != 1 {
		t.Fatalf("death-effects-happy loss not recorded: %#v", next.Nations[1].PeaceMovement)
	}
}

func TestRetailPlantNukeRequiresNuclearAdvanceWhenRuntimeDeclaresIt(t *testing.T) {
	state := task11State()
	state.Cities[0].Population = 8
	state.Units[0].Effects = []Effect{{ID: "nuke", Kind: EffectSpecialPlantNuke}}
	runtime := rules.RuntimeRules{Advances: []rules.RuntimeAdvanceDefinition{{ID: "ADVANCE_NUCLEAR_POWER", Cost: 1, AllowsPlantNuke: true}}}
	_, _, err := plantNuke(&state, &state.Units[0], &state.Cities[0], &state.Nations[0], &state.Nations[1], fixtureSources(), &runtime)
	if err == nil || !strings.Contains(err.Error(), "special action") {
		t.Fatalf("plant nuke without advance err=%v", err)
	}

	state.Nations[0].Advances = []string{"ADVANCE_NUCLEAR_POWER"}
	eventType, _, err := plantNuke(&state, &state.Units[0], &state.Cities[0], &state.Nations[0], &state.Nations[1], fixtureSources(), &runtime)
	if err != nil {
		t.Fatal(err)
	}
	if state.Cities[0].Population != 4 || eventType != EventNukePlanted {
		t.Fatalf("plant nuke with advance outcome city=%+v event=%s", state.Cities[0], eventType)
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
	state.Units[0].Effects = []Effect{{ID: "nuke", Kind: EffectSpecialPlantNuke, Value: 10}}
	next, events, err = task11Apply(t, state, CommandPlantNuke, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].Population != 8 || next.Tiles[1].Pollution != 0 || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventNukeFailed {
		t.Fatalf("chance-failed nuke outcome city=%+v tile=%+v unit=%+v events=%+v", next.Cities[0], next.Tiles[1], next.Units[0], events)
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

	failedState := task11State()
	failedState.Cities[0].Population = 9
	failedState.Units[0].Effects = []Effect{{ID: "investigate", Kind: EffectSpecialInvestigateCity, Value: 10}}
	failed, failedEvents, err := task11Apply(t, failedState, CommandInvestigateCity, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if len(failed.Nations[0].CityReports) != 0 || failed.Units[0].Movement != 3 || len(failedEvents) != 1 || failedEvents[0].Type != EventCityInvestigationFailed {
		t.Fatalf("chance-failed investigation outcome nation=%+v unit=%+v events=%+v", failed.Nations[0], failed.Units[0], failedEvents)
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

func TestRetailSlaveUprisingConsumesSlaveAndRiotMarksCity(t *testing.T) {
	state := task11State()
	state.Cities[0].Population = 5
	state.Cities[0].Specialists = map[string]int{"POP_SLAVE": 2}
	state.Units[0].Effects = []Effect{{ID: "uprising", Kind: EffectSpecialSlaveUprising}}

	next, events, err := task11Apply(t, state, CommandSlaveUprising, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].Specialists["POP_SLAVE"] != 1 || !next.Cities[0].Rioting || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventSlaveUprising || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("slave uprising state/events = city %+v unit %+v events %#v", next.Cities[0], next.Units[0], events)
	}
}

func TestRetailIndulgenceSalesTransferGoldAndTimedHappiness(t *testing.T) {
	state := task11State()
	state.Nations[0].Gold = 1
	state.Nations[1].Gold = 10
	state.Units[0].Effects = []Effect{{ID: "indulgence", Kind: EffectSpecialIndulgenceSales}}
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Constants: rules.RuntimeConstants{UnconvertedIndulgenceGold: 5, UnconvertedIndulgenceHappiness: 4}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}}}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}

	next, events, err := task11Apply(t, state, CommandIndulgenceSales, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Nations[0].Gold != 6 || next.Nations[1].Gold != 5 || next.Units[0].Movement != 3 || cityHappiness(next, next.Cities[0]) != 79 || !hasTimedHappinessAttack(next.Cities[0], 4, 1) || len(events) != 1 || events[0].Type != EventIndulgenceSold || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("indulgence state/events = nations %+v city %+v unit %+v events %#v", next.Nations, next.Cities[0], next.Units[0], events)
	}
}

func TestRetailInciteRevolutionUsesChanceAndChangesCityOwner(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = []Effect{{ID: "incite", Kind: EffectSpecialInciteRevolution, Value: 100}, {ID: "elite", Kind: EffectSpecialInciteRevolutionEliteChance, Value: 100}, {ID: "death", Kind: EffectSpecialInciteRevolutionDeathChance, Value: 0}}

	next, events, err := task11Apply(t, state, CommandInciteRevolution, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].OwnerID != "player-a" || next.Units[0].Movement != 3 || !hasEffect(next.Cities[0].Effects, EffectWatchful) || len(events) != 1 || events[0].Type != EventRevolutionIncited || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("incite revolution state/events = city %+v unit %+v events %#v", next.Cities[0], next.Units[0], events)
	}
}

func TestRetailCauseUnhappinessAddsTimedHappinessAttack(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = []Effect{{ID: "cause", Kind: EffectSpecialCauseUnhappiness, Value: 100}, {ID: "timer", Kind: EffectSpecialCauseUnhappinessTimer, Value: 2}, {ID: "amount", Kind: EffectSpecialCauseUnhappinessAmount, Value: 5}}

	next, events, err := task11Apply(t, state, CommandCauseUnhappiness, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventCauseUnhappinessApplied || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("cause unhappiness event = %#v", events)
	}
	city := next.Cities[0]
	if next.Units[0].Movement != 3 || cityHappiness(next, city) != 70 || !hasTimedHappinessAttack(city, -5, 2) || !hasEffect(city.Effects, EffectWatchful) {
		t.Fatalf("cause unhappiness state = unit %+v city %+v happiness %d", next.Units[0], city, cityHappiness(next, city))
	}
	advanceTemporaryCityEffects(&city)
	if !hasTimedHappinessAttack(city, -5, 1) {
		t.Fatalf("happiness attack did not tick: %+v", city.Effects)
	}
	advanceTemporaryCityEffects(&city)
	if hasTimedHappinessAttack(city, -5, 1) || hasTimedHappinessAttack(city, -5, 2) {
		t.Fatalf("happiness attack did not expire: %+v", city.Effects)
	}
}

func hasTimedHappinessAttack(city City, amount int, turns int) bool {
	for _, effect := range city.Effects {
		if effect.Kind == EffectHappiness && effect.Value == amount && effect.Turns == turns {
			return true
		}
	}
	return false
}

func TestRetailCreateFranchiseMarksCityOnce(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = []Effect{{ID: "franchise", Kind: EffectSpecialCreateFranchise}}
	next, events, err := task11Apply(t, state, CommandCreateFranchise, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Units[0].Movement != 3 || !hasEffect(next.Cities[0].Effects, EffectFranchise) || !hasEffect(next.Cities[0].Effects, EffectWatchful) || len(events) != 1 || events[0].Type != EventFranchiseCreated || strings.Contains(string(events[0].Payload), "target-city") {
		t.Fatalf("franchise outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
	}

	state = task11State()
	state.Units[0].Effects = []Effect{{ID: "franchise", Kind: EffectSpecialCreateFranchise, Value: 10}}
	failed, failedEvents, err := task11Apply(t, state, CommandCreateFranchise, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if !hasEffect(failed.Cities[0].Effects, EffectWatchful) || failed.Units[0].Movement != 3 || len(failedEvents) != 1 || failedEvents[0].Type != EventFranchiseFailed {
		t.Fatalf("chance-failed franchise outcome city=%+v unit=%+v events=%+v", failed.Cities[0], failed.Units[0], failedEvents)
	}

	next, _, err = task11Apply(t, next, CommandCreateFranchise, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err == nil {
		t.Fatalf("duplicate franchise accepted: %+v", next.Cities[0].Effects)
	}
}

func TestRetailProhibitSlaversBlocksSlaveRaidCommands(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = []Effect{{ID: "slave", Kind: EffectSpecialSlaveRaid}}
	state.Wonders = []Wonder{{ID: "WONDER_EMANCIPATION", OwnerID: "player-a", CityID: "target-city", Effects: []Effect{{ID: "prohibit", Kind: EffectProhibitSlavers}}}}
	if _, _, err := task11Apply(t, state, CommandSlaveRaid, SpecialActionPayload{UnitID: "agent", CityID: "target-city"}); err == nil || !strings.Contains(err.Error(), "special action") {
		t.Fatalf("slave raid allowed under prohibit-slavers wonder err=%v", err)
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
			state.Units[0].Effects = []Effect{{ID: tc.name, Kind: tc.effect, Value: 10}}
			next, events, err = task11Apply(t, state, tc.kind, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
			if err != nil {
				t.Fatal(err)
			}
			if next.Cities[0].Population != 3 || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventSlaveRaidFailed {
				t.Fatalf("chance-failed slave raid outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
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

			state = task11State()
			state.Cities[0].Population = 5
			state.Units[0].Effects = []Effect{{ID: tc.name, Kind: tc.effectKind, Value: 10}}
			next, events, err = task11Apply(t, state, tc.command, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
			if err != nil {
				t.Fatal(err)
			}
			if next.Cities[0].Population != 5 || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventTerrorFailed {
				t.Fatalf("chance-failed terror outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
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
	state.Units[0].Effects = []Effect{{ID: "convert", Kind: EffectSpecialConvertCity, Value: 10}}
	next, events, err = task11Apply(t, state, CommandConvertCity, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].OwnerID != "player-b" || next.Units[0].Movement != 3 || len(events) != 1 || events[0].Type != EventCityConversionFailed {
		t.Fatalf("chance-failed conversion outcome city=%+v unit=%+v events=%+v", next.Cities[0], next.Units[0], events)
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

func TestRetailPreventConversionEmpireWonderBlocksCityConversion(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = []Effect{{ID: "convert", Kind: EffectSpecialConvertCity}}
	state.Wonders = []Wonder{{ID: "WONDER_GUTTENBERGS_BIBLE", OwnerID: "player-b", CityID: "target-city", Effects: []Effect{{ID: "prevent", Kind: EffectPreventConversionEmpire}}}}
	next, events, err := task11Apply(t, state, CommandConvertCity, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].OwnerID != "player-b" || len(events) != 1 || events[0].Type != EventCityConversionFailed {
		t.Fatalf("conversion with empire prevention = city:%#v events:%#v", next.Cities[0], events)
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

func TestRetailSpiesEverywhereStopsTechnologyTheftAndKillsAgent(t *testing.T) {
	state := task11State()
	state.Units[0].Effects = []Effect{{ID: "steal", Kind: EffectSpecialSteal, Value: 100}}
	state.Wonders = []Wonder{{ID: "WONDER_THE_AGENCY", OwnerID: "player-b", CityID: "target-city", Effects: []Effect{{ID: "spies", Kind: EffectSpiesEverywhere}}}}
	next, events, err := task11Apply(t, state, CommandStealTechnology, SpecialActionPayload{UnitID: "agent", CityID: "target-city"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := unitByID(next, "agent"); ok || len(events) != 1 || events[0].Type != EventTechnologyFailed || containsString(next.Nations[0].Advances, "alpha") || containsString(next.Nations[0].Advances, "zeta") {
		t.Fatalf("spies-everywhere theft result units=%#v nations=%#v events=%#v", next.Units, next.Nations, events)
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
