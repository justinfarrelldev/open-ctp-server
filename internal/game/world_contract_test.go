package game

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerationConsumesAllHostParams(t *testing.T) {
	game := &Game{
		ID:               44,
		Style:            "standard_ctp2",
		RulesJSON:        `["fast"]`,
		ExclusionsJSON:   `{"units":["settler"]}`,
		StartingSettlers: 2,
		StartingGold:     250,
	}
	players := []GamePlayer{{AccountID: 7, Country: "Greeks"}, {AccountID: 8, Country: "Romans"}}

	input, err := NormalizeWorldGenerationInput(game, players)
	if err != nil {
		t.Errorf("NormalizeWorldGenerationInput() error = %v", err)
	}
	if input.GameID != 44 {
		t.Errorf("got game_id %d want 44", input.GameID)
	}
	if input.HostParams.Style != game.Style {
		t.Errorf("got style %q want %q", input.HostParams.Style, game.Style)
	}
	if input.HostParams.RulesJSON != game.RulesJSON {
		t.Errorf("got rules_json %q want %q", input.HostParams.RulesJSON, game.RulesJSON)
	}
	if input.HostParams.ExclusionsJSON != game.ExclusionsJSON {
		t.Errorf("got exclusions_json %q want %q", input.HostParams.ExclusionsJSON, game.ExclusionsJSON)
	}
	if input.HostParams.StartingSettlers != game.StartingSettlers {
		t.Errorf("got starting_settlers %d want %d", input.HostParams.StartingSettlers, game.StartingSettlers)
	}
	if input.HostParams.StartingGold != game.StartingGold {
		t.Errorf("got starting_gold %d want %d", input.HostParams.StartingGold, game.StartingGold)
	}
	if len(input.PlayerOrder) != 2 || input.PlayerOrder[0] != 7 || input.PlayerOrder[1] != 8 {
		t.Errorf("got player_order %+v want [7 8]", input.PlayerOrder)
	}
	if input.HostParams.Seed == nil {
		t.Error("missing required host param: seed")
	}
	if input.HostParams.Wetness == nil {
		t.Error("missing required host param: wetness")
	}
	if strings.TrimSpace(input.HostParams.MapSize) == "" {
		t.Error("missing required host param: map_size")
	}
	if strings.TrimSpace(input.HostParams.Climate) == "" {
		t.Error("missing required host param: climate")
	}
	if strings.TrimSpace(input.GeneratorVersion) == "" && strings.TrimSpace(input.HostParams.GeneratorVersion) == "" {
		t.Error("missing required host param: generator_version")
	}
}

func TestWorldStartContract(t *testing.T) {
	t.Run("rejects duplicate world generation for the same game", func(t *testing.T) {
		game := &Game{ID: 44, Status: DefaultGameStatus}
		err := ValidateWorldStartContract(game, true)
		if err == nil {
			t.Fatal("expected duplicate world start conflict")
		}
		want := "world already exists for game 44"
		if err.Error() != want {
			t.Fatalf("got error %q want %q", err.Error(), want)
		}
	})

	t.Run("rejects non setup games before world generation begins", func(t *testing.T) {
		game := &Game{ID: 44, Status: "started"}
		err := ValidateWorldStartContract(game, false)
		if err == nil {
			t.Fatal("expected invalid start conflict")
		}
		want := fmt.Sprintf("game %d is not in setup", game.ID)
		if err.Error() != want {
			t.Fatalf("got error %q want %q", err.Error(), want)
		}
	})

	t.Run("rejects ended games before attempting any new world start", func(t *testing.T) {
		game := &Game{ID: 44, Status: "ended"}
		err := ValidateWorldStartContract(game, false)
		if err == nil {
			t.Fatal("expected ended-game start conflict")
		}
		want := fmt.Sprintf("game %d is not in setup", game.ID)
		if err.Error() != want {
			t.Fatalf("got error %q want %q", err.Error(), want)
		}
	})
}

func TestWorldRouteSurfaceContract(t *testing.T) {
	mainPath := filepath.Join("..", "..", "main.go")
	content, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", mainPath, err)
	}
	text := string(content)

	if strings.Contains(text, `"/game/chat"`) {
		t.Fatal("/game/chat route must not be registered; active games continue to use lobby chat")
	}
	if strings.Contains(text, `"/world/chat"`) {
		t.Fatal("/world/chat route must not be registered; active games continue to use lobby chat")
	}
	if !strings.Contains(text, `"/lobby/send_chat"`) {
		t.Fatal("expected lobby chat route to remain registered as the active-game chat surface")
	}
}
