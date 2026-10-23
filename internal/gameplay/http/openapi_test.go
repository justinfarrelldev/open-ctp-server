package gameplayhttp

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestOpenAPIDocumentsGameplayCommandRoute(t *testing.T) {
	contents, err := os.ReadFile("../../../docs/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths       map[string]json.RawMessage `json:"paths"`
		Definitions map[string]json.RawMessage `json:"definitions"`
	}
	if err := json.Unmarshal(contents, &document); err != nil {
		t.Fatal(err)
	}
	path, ok := document.Paths["/game/{game_id}/{command}"]
	if !ok || !strings.Contains(string(path), "post") || !strings.Contains(string(path), "429") {
		t.Fatalf("gameplay command route or rate-limit response is undocumented: %s", path)
	}
	for _, name := range []string{"gameplayhttp.CommandRequestDocs", "gameplayhttp.CommandResponseDocs", "service.Rejection"} {
		if _, ok := document.Definitions[name]; !ok {
			t.Fatalf("missing gameplay OpenAPI schema %q", name)
		}
	}
}
