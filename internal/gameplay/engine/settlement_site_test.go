package engine

import (
	"bytes"
	"encoding/json"
	"testing"
)

// The national owner may differ from the founder. UnitData::Settle checks city
// influence instead; Player::CreateCity separately rejects occupied/adjacent sites.
func TestSettlementSiteParityAcrossRulesPins(t *testing.T) {
	for _, configured := range []bool{false, true} {
		pinName := "legacy"
		if configured {
			pinName = "territory-configured"
		}
		t.Run(pinName, func(t *testing.T) {
			for _, tc := range []struct {
				name, owner, influence, cityTile, cityOwner string
				allowed                                     bool
			}{
				{name: "neutral", allowed: true},
				{name: "own national", owner: "player-a", allowed: true},
				{name: "foreign national", owner: "player-b", allowed: true},
				{name: "own influence", owner: "player-a", influence: "existing", cityTile: "far", cityOwner: "player-a"},
				{name: "foreign influence", owner: "player-b", influence: "existing", cityTile: "far", cityOwner: "player-b"},
				{name: "occupied", cityTile: "site", cityOwner: "player-b"},
				{name: "adjacent", cityTile: "near", cityOwner: "player-b"},
				{name: "diagonally adjacent own city", cityTile: "diagonal", cityOwner: "player-a"},
				{name: "foreign national outside distant city", owner: "player-b", cityTile: "far", cityOwner: "player-b", allowed: true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					state := task10State()
					if configured {
						territory, _ := territoryState(t)
						state.Rules = territory.Rules
					}
					state.Tiles = []Tile{
						{ID: "site", Terrain: TerrainPlains, Open: true, OwnerID: tc.owner, InfluenceCityID: tc.influence},
						{ID: "near", X: 1, Terrain: TerrainPlains, Open: true},
						{ID: "diagonal", X: 1, Y: 1, Terrain: TerrainPlains, Open: true},
						{ID: "far", X: 3, Terrain: TerrainPlains, Open: true},
					}
					state.Units = state.Units[:1]
					state.Units[0].TileID = "site"
					if tc.cityTile != "" {
						state.Cities = []City{{ID: "existing", OwnerID: tc.cityOwner, TileID: tc.cityTile, Name: "Existing", Population: 1}}
					}
					before, err := CanonicalStateJSON(state)
					if err != nil {
						t.Fatal(err)
					}
					command := Command{SchemaVersion: CommandSchemaV1, ID: "found", ActorPlayerID: "player-a", Type: CommandFoundCity, Payload: []byte(`{"unit_id":"settler-a","city_id":"new-city","name":"Athens"}`)}
					next, events, _, err := Apply(state, command, state.Rules, fixtureSources())
					after, marshalErr := CanonicalStateJSON(state)
					if marshalErr != nil || !bytes.Equal(before, after) {
						t.Fatal("founding mutated caller state")
					}
					if !tc.allowed {
						if err == nil || len(events) != 0 {
							t.Fatalf("illegal site accepted: err=%v events=%+v", err, events)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					city, ok := cityByID(next, "new-city")
					if !ok || city.TileID != "site" || city.OwnerID != "player-a" || city.Name != "Athens" || len(next.Units) != 0 || len(events) != 1 || events[0].Type != "city.founded" || next.Revision != state.Revision+1 {
						t.Fatalf("non-atomic founding: city=%+v units=%+v events=%+v", city, next.Units, events)
					}
					tile, _ := tileByID(next, "site")
					if tile.OwnerID != "player-a" || tile.InfluenceCityID != "new-city" {
						t.Fatalf("center territory not initialized: %+v", tile)
					}
					if err := CheckInvariants(next); err != nil {
						t.Fatal(err)
					}
					replay, replayEvents, _, err := Apply(state, command, state.Rules, fixtureSources())
					if err != nil || artifactHash(t, next, events) != artifactHash(t, replay, replayEvents) {
						t.Fatalf("founding replay changed: %v", err)
					}
					raw, err := CanonicalStateJSON(next)
					if err != nil {
						t.Fatal(err)
					}
					restored, err := DecodeCanonicalStateJSON(raw)
					if err != nil || artifactHash(t, next, nil) != artifactHash(t, restored, nil) {
						t.Fatalf("founding restart changed: %v", err)
					}
					command.ID = "duplicate"
					if _, _, _, err := Apply(restored, command, restored.Rules, fixtureSources()); err == nil {
						t.Fatal("consumed settler/duplicate city accepted")
					}
				})
			}
		})
	}
}

func TestSettlementRetainsUnitIdentityAndTurnGuards(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*State, *FoundCityPayload, *Command)
	}{
		{"nonsettler", func(s *State, _ *FoundCityPayload, _ *Command) { s.Units[0].Effects = nil }},
		{"foreign unit", func(s *State, _ *FoundCityPayload, _ *Command) { s.Units[0].OwnerID = "player-b" }},
		{"army member", func(s *State, _ *FoundCityPayload, _ *Command) {
			s.Units[0].ArmyID = "army"
			s.Armies = []Army{{ID: "army", OwnerID: "player-a", UnitIDs: []string{"settler-a"}}}
		}},
		{"transported", func(s *State, _ *FoundCityPayload, _ *Command) {
			s.Units[0].TransportID = "settler-b"
			s.Units[0].TileID = ""
			s.Units[1].OwnerID = "player-a"
			s.Units[1].CargoCapacity = 1
			s.Units[1].CargoDomains = []string{DomainLand}
		}},
		{"missing unit", func(_ *State, p *FoundCityPayload, _ *Command) { p.UnitID = "missing" }},
		{"empty identity", func(_ *State, p *FoundCityPayload, _ *Command) { p.CityID = "" }},
		{"duplicate identity", func(s *State, p *FoundCityPayload, _ *Command) {
			s.Cities = []City{{ID: p.CityID, OwnerID: "player-b", TileID: "c"}}
		}},
		{"invalid name", func(_ *State, p *FoundCityPayload, _ *Command) { p.Name = "\n" }},
		{"inactive player", func(_ *State, _ *FoundCityPayload, c *Command) { c.ActorPlayerID = "player-b" }},
		{"terminal", func(s *State, _ *FoundCityPayload, _ *Command) {
			s.Terminal = TerminalState{Ended: true, Reason: "conquest", WinningPlayerID: "player-b"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := task10State()
			state.Tiles[0].OwnerID = "player-b"
			payload := FoundCityPayload{UnitID: "settler-a", CityID: "new-city", Name: "Athens"}
			command := Command{SchemaVersion: CommandSchemaV1, ID: "found", ActorPlayerID: "player-a", Type: CommandFoundCity}
			tc.mutate(&state, &payload, &command)
			if err := ValidateState(state); err != nil {
				t.Fatalf("invalid guard fixture: %v", err)
			}
			command.Payload, _ = json.Marshal(payload)
			before, _ := json.Marshal(state)
			_, events, _, err := Apply(state, command, state.Rules, fixtureSources())
			after, _ := json.Marshal(state)
			if err == nil || len(events) != 0 || !bytes.Equal(before, after) {
				t.Fatalf("guard failed: err=%v events=%+v", err, events)
			}
		})
	}
}
