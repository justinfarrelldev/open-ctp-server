package gameplayhttp

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

// These formerly advertised routes must traverse the same authenticated gateway
// as ordinary commands, with no client authority or relaxed unknown-field policy.
func TestClientReportedRoutesDispatchExactPayloads(t *testing.T) {
	cases := []struct{ route, kind, payload string }{
		{"assassinate-ruler", engine.CommandAssassinateRuler, `{"unit_id":"agent","city_id":"city"}`},
		{"bio-terror", engine.CommandBioTerror, `{"unit_id":"agent","city_id":"city"}`},
		{"bombard", engine.CommandBombard, `{"unit_id":"agent","target_tile_id":"tile"}`},
		{"convert-city", engine.CommandConvertCity, `{"unit_id":"agent","city_id":"city"}`},
		{"create-franchise", engine.CommandCreateFranchise, `{"unit_id":"agent","city_id":"city"}`},
		{"create-park", engine.CommandCreatePark, `{"unit_id":"agent","tile_id":"tile"}`},
		{"expel", engine.CommandExpel, `{"unit_id":"agent","target_unit_id":"intruder"}`},
		{"injoin", engine.CommandInjoin, `{"unit_id":"agent","city_id":"city"}`},
		{"investigate-city", engine.CommandInvestigateCity, `{"unit_id":"agent","city_id":"city"}`},
		{"nano-terror", engine.CommandNanoTerror, `{"unit_id":"agent","city_id":"city"}`},
		{"plague", engine.CommandPlague, `{"unit_id":"agent","city_id":"city"}`},
		{"plant-nuke", engine.CommandPlantNuke, `{"unit_id":"agent","city_id":"city"}`},
		{"reform", engine.CommandReform, `{"unit_id":"agent","city_id":"city"}`},
		{"settler-slave-raid", engine.CommandSettlerSlaveRaid, `{"unit_id":"agent","city_id":"city"}`},
		{"slave-raid", engine.CommandSlaveRaid, `{"unit_id":"agent","city_id":"city"}`},
		{"sue", engine.CommandSue, `{"unit_id":"agent","target_unit_id":"intruder"}`},
		{"sue", engine.CommandSue, `{"unit_id":"agent","city_id":"city"}`},
		{"underground-railway", engine.CommandUndergroundRailway, `{"unit_id":"agent","city_id":"city"}`},
		{"upgrade-unit", engine.CommandUpgradeUnit, `{"unit_id":"agent","upgrade_to":"warrior"}`},
		{"pillage", engine.CommandPillage, `{"unit_id":"agent"}`},
	}
	for _, tc := range cases {
		t.Run(tc.route+tc.payload, func(t *testing.T) {
			for _, forged := range []bool{false, true} {
				executor := &recordingExecutor{}
				body := `{"session_id":1,"expected_revision":2,"idempotency_key":"key","command_id":"command",` + tc.payload[1:]
				if forged {
					body = strings.TrimSuffix(body, "}") + `,"actor_player_id":"victim"}`
				}
				response := httptest.NewRecorder()
				authenticatedHandler(executor).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/game/9/"+tc.route, strings.NewReader(body)))
				if forged {
					if response.Code != http.StatusBadRequest || executor.calls != 0 {
						t.Fatalf("forged field reached executor: %d calls=%d", response.Code, executor.calls)
					}
					continue
				}
				if response.Code != http.StatusOK || executor.calls != 1 || executor.request.Type != tc.kind {
					t.Fatalf("route failed: %d calls=%d request=%+v", response.Code, executor.calls, executor.request)
				}
				var want, got map[string]any
				if err := json.Unmarshal([]byte(tc.payload), &want); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(executor.request.Payload, &got); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("payload changed: got=%v want=%v", got, want)
				}
			}
		})
	}
}

// Inspect every literal action declaration, including branches that require rare
// retail abilities. Runtime eligibility fixtures alone cannot detect route drift.
func TestAdvertisedLiteralRoutesAndFieldsAreRegistered(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "../visibility/visibility.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		name, ok := call.Fun.(*ast.Ident)
		if !ok || name.Name != "action" || len(call.Args) < 6 {
			return true
		}
		route, ok := call.Args[1].(*ast.BasicLit)
		if !ok || route.Kind != token.STRING {
			return true
		}
		path, err := strconv.Unquote(route.Value)
		if err != nil {
			t.Fatal(err)
		}
		definition, ok := commands[path]
		if !ok {
			t.Errorf("advertised route %q is unregistered", path)
			return true
		}
		count++
		ast.Inspect(call.Args[5], func(fieldNode ast.Node) bool {
			literal, ok := fieldNode.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			field, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !definition.allowed[field] {
				t.Errorf("route %q rejects advertised field %q", path, field)
			}
			return true
		})
		return true
	})
	if count < 30 {
		t.Fatalf("route drift guard inspected too few declarations: %d", count)
	}
}
