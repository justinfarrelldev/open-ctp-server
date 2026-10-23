package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestRetailStartupPinsSelectedCivilizationsAndOwnFoundingSuggestions(t *testing.T) {
	t.Setenv(retailRulesModeEnv, "retail")
	t.Setenv(retailGameDataEnv, startingAdvanceGameDataRoot(t))
	pin, err := serverRuntimePin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := initialGameplayState(startingAdvanceSnapshot(), "sequential", pin, []GamePlayer{{AccountID: 7, Country: "not-a-source-civilization"}, {AccountID: 8, Country: "Greeks", Controller: engine.ControllerAI}}); !errors.Is(err, errCivilizationSetup) {
		t.Fatalf("unknown retail civilization did not fail safely: %v", err)
	}
	state, err := initialGameplayState(startingAdvanceSnapshot(), "sequential", pin, []GamePlayer{{AccountID: 7, Country: "Romans"}, {AccountID: 8, Country: "Greeks", Controller: engine.ControllerAI}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ actor, name, civ string }{{"account-7", "Rome", "ROMAN"}, {"account-8", "Athens", "GREEK"}} {
		p, err := visibility.Project(state, tc.actor)
		if err != nil {
			t.Fatal(err)
		}
		inspector, err := engine.NewInspector(state)
		if err != nil {
			t.Fatal(err)
		}
		if got := inspector.SuggestedCityName(tc.actor); got != tc.name {
			t.Fatalf("%s name = %q, want %s", tc.actor, got, tc.name)
		}
		if tc.actor == "account-7" {
			found := false
			for _, a := range p.LegalActions {
				if a.Command == engine.CommandFoundCity {
					var payload engine.FoundCityPayload
					if err := json.Unmarshal(a.Payload, &payload); err != nil {
						t.Fatal(err)
					}
					if payload.Name != tc.name {
						t.Fatalf("projected founding name = %s", payload.Name)
					}
					found = true
				}
			}
			if !found {
				t.Fatal("no projected founding action")
			}
		}
		for _, player := range state.Players {
			if player.ID == tc.actor {
				for _, nation := range state.Nations {
					if nation.ID == player.NationID && nation.CivilizationID != tc.civ {
						t.Fatalf("selected civilization lost: %+v", nation)
					}
				}
			}
		}
	}
}
