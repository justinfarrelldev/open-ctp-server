package visibility

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestObservedRiverEdgesPersistWithoutProjectingHiddenNeighbors(t *testing.T) {
	state := clientMetadataState()
	// The scout at t0 sees t2, but not t3. The tile artwork at t2 has
	// NW/SE continuations even though the SE tile remains unprojected.
	state.Tiles[2].River = true
	state.Tiles[3].River = true
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	wire := metadataWire(t, metadataTile(t, p, "t2"))
	geometry, ok := wire["river_geometry"].(map[string]any)
	if !ok || geometry["version"] != "ctp2-river-edges-v1" || geometry["edges"] != float64(10) || geometry["observed_revision"] != float64(state.Revision) {
		t.Fatalf("missing observed art fact: %v", wire)
	}
	for _, tile := range p.Tiles {
		if tile.ID == "t3" {
			t.Fatal("hidden neighbor leaked")
		}
	}
	original, _ := json.Marshal(geometry)
	state.Units[0].TileID = "t10"
	state.Tiles[3].River = false
	state.Tiles[2].River = false
	state.Revision++
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	state, err = engine.DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	p, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	remembered, _ := json.Marshal(metadataWire(t, metadataTile(t, p, "t2"))["river_geometry"])
	if string(original) != string(remembered) {
		t.Fatalf("hidden changes altered remembered art %s -> %s", original, remembered)
	}
	state.Units[0].TileID = "t0"
	p, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := metadataWire(t, metadataTile(t, p, "t2"))["river_geometry"]; ok {
		t.Fatal("nonriver has river art")
	}
	// A legacy remembered observation remains unknown; don't backfill via authority.
	state.Units[0].TileID = "t10"
	for n := range state.Fog {
		state.Fog[n].Inspection = nil
	}
	p, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := metadataWire(t, metadataTile(t, p, "t2"))["river_geometry"]; ok {
		t.Fatal("legacy unknown art backfilled")
	}
}
