package ai

import (
	"context"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestLegacyClosedAndQueuedProjectionKeepMetadataOnlyCloseout(t *testing.T) {
	for _, status := range []string{"closed", "queued"} {
		source := staticProjection{bytes: []byte(`{"revision":4,"planning_status":"` + status + `","units":"unused legacy field"}`)}
		gateway := &recordingGateway{revision: 4}
		driver := NewDriver(source, gateway, Config{})
		if _, err := driver.Drive(context.Background(), 44, "7", 4); err != nil {
			t.Fatal(err)
		}
		if status == "closed" && len(gateway.requests) != 0 {
			t.Fatal("closed actor submitted")
		}
		if status == "queued" && (len(gateway.requests) != 1 || gateway.requests[0].Type != engine.CommandRoundComplete) {
			t.Fatal("queued closeout changed")
		}
	}
}
