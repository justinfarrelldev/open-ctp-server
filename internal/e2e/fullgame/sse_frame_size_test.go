package fullgame

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestPublicExplorationAndResumeAcceptLargeViewerSSEFrames(t *testing.T) {
	empty := ""
	p := visibility.Projection{Revision: 1}
	for i := 0; i < 1200; i++ {
		p.Tiles = append(p.Tiles, visibility.TileProjection{ID: fmt.Sprintf("observed-%d", i), Terrain: "plains", OwnerID: &empty, InfluenceCityID: &empty})
	}
	snapshot, _ := json.Marshal(p)
	if len(snapshot) <= 64<<10 {
		t.Fatal("fixture must exceed Go Scanner's default token size")
	}
	frame, _ := json.Marshal(map[string]any{"current_revision": 1, "snapshot": json.RawMessage(snapshot)})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "id: 1\nevent: resync\ndata: %s\n\n", frame)
	}))
	defer server.Close()
	transport := &HTTPExplorationTransport{BaseURL: server.URL, Client: server.Client(), GameID: 1, SessionID: 1}
	if _, data, err := transport.Bootstrap(context.Background(), 0); err != nil || string(data) != string(snapshot) {
		t.Fatalf("large current SSE read failed: %v", err)
	}
	if _, data, err := scanPublicEvents(context.Background(), transport, 0, false); err != nil || string(data) != string(frame) {
		t.Fatalf("large resume SSE read failed: %v", err)
	}
}
