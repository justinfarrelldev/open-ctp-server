package fullgame

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type observedConquestExecutor struct{ artifacts []Artifacts }

func (e *observedConquestExecutor) ExecutePublicConquest(_ context.Context, _ ConquestTemplate, runID string) (Artifacts, error) {
	if runID == "run-1" {
		return e.artifacts[0], nil
	}
	return e.artifacts[1], nil
}

func TestFixedConquestTemplateBuildsRoleBoundPlannerSchedule(t *testing.T) {
	projection, err := json.Marshal(visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "start", X: 0, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "approach", X: 1, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		},
		Units: []visibility.UnitProjection{{ID: "attacker-unit", OwnerID: "account-1", TileID: "start", Domain: "land", Movement: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	schedule, err := FixedConquestTemplate().Schedule(projection, "attacker", "attacker-unit", "approach", "defender-army")
	if err != nil {
		t.Fatal(err)
	}
	if len(schedule) != 4 || schedule[0].Command != "army-create" || schedule[1].Command != "move" || schedule[2].Command != "combat" || schedule[3].Command != "pass" {
		t.Fatalf("role-bound conquest schedule = %#v", schedule)
	}
	if schedule[1].Role != "attacker" || string(schedule[1].Payload) != `{"path":["approach"],"unit_id":"attacker-unit"}` {
		t.Fatalf("planner move step = %#v", schedule[1])
	}
}

func TestFixedConquestTemplateRequiresVisibleDiscoveryBeforePlanningCombat(t *testing.T) {
	before, err := json.Marshal(visibility.Projection{Tiles: []visibility.TileProjection{{ID: "start", X: 0, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}}}, Units: []visibility.UnitProjection{{ID: "attacker-unit", OwnerID: "account-1", TileID: "start", Domain: "land", Movement: 2}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := FixedConquestTemplate().Schedule(before, "attacker", "attacker-unit", "approach", "defender-army"); err == nil {
		t.Fatal("conquest template planned an unseen defender route")
	}
	after, err := json.Marshal(visibility.Projection{Tiles: []visibility.TileProjection{{ID: "start", X: 0, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}}, {ID: "approach", X: 1, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}}, {ID: "defender", X: 2, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}}}, Units: []visibility.UnitProjection{{ID: "attacker-unit", OwnerID: "account-1", TileID: "start", Domain: "land", Movement: 2}, {ID: "visible-defender", OwnerID: "account-2", TileID: "defender"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := FixedConquestTemplate().Schedule(after, "attacker", "attacker-unit", "approach", "defender-army"); err != nil {
		t.Fatalf("conquest template rejected a legally visible defender route: %v", err)
	}
}

func TestObserveConquestTwiceRetainsOnlyEqualNaturalConquest(t *testing.T) {
	artifacts := Artifacts{TerminalWinner: "account-1", TerminalReason: "conquest", StateHash: "state", EventHash: "events"}
	observed, err := ObserveConquestTwice(context.Background(), FixedConquestTemplate(), &observedConquestExecutor{artifacts: []Artifacts{artifacts, artifacts}})
	if err != nil {
		t.Fatal(err)
	}
	if observed.First != artifacts || observed.Second != artifacts {
		t.Fatalf("observed artifacts = %#v", observed)
	}
	mismatch := artifacts
	mismatch.EventHash = "different"
	if _, err := ObserveConquestTwice(context.Background(), FixedConquestTemplate(), &observedConquestExecutor{artifacts: []Artifacts{artifacts, mismatch}}); err == nil {
		t.Fatal("observation gate retained unequal artifacts")
	}
}
