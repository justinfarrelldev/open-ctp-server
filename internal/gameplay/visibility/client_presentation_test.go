package visibility

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestClientPresentationProjectsOnlyOwnedHistoricalResults(t *testing.T) {
	state := clientMetadataState()
	state.Nations = []engine.Nation{{ID: "a", Statistics: []engine.StatisticsSample{{Turn: 1, Score: 42, CityCount: 1}}}, {ID: "b", Statistics: []engine.StatisticsSample{{Turn: 1, Score: 999}}}}
	state.Notifications = []engine.Notification{{ID: "owned-result", PlayerID: "a", Kind: "production.completed", Turn: 1, CityID: "old-home", TileID: "t0", ContentID: "granary"}, {ID: "HIDDEN_RESULT_TOKEN", PlayerID: "b", Kind: "production.completed", Turn: 1, CityID: "hidden-city", TileID: "t11", ContentID: "secret-content"}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	wire := metadataWire(t, projection)
	notices, ok := wire["notifications"].([]any)
	if !ok || len(notices) != 1 {
		t.Errorf("owner result missing: %v", wire["notifications"])
	}
	statistics, ok := wire["statistics"].([]any)
	if !ok || len(statistics) != 1 || statistics[0].(map[string]any)["score"] != float64(42) {
		t.Errorf("owner history missing: %v", wire["statistics"])
	}
	raw, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("HIDDEN_RESULT_TOKEN")) || bytes.Contains(raw, []byte("secret-content")) || bytes.Contains(raw, []byte("hidden-city")) {
		t.Fatalf("foreign result/location leaked: %s", raw)
	}
	for _, tile := range projection.Tiles {
		if tile.ID == "t11" {
			t.Fatal("history revealed a never-discovered location")
		}
	}
}
