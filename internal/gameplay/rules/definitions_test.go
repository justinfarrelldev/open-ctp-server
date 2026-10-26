package rules

import (
	"encoding/json"
	"testing"
)

func TestDefinitionMetadataAndCompactionMatchJSONCodecs(t *testing.T) {
	for _, raw := range []json.RawMessage{nil, []byte(`null`), []byte(`{}`), []byte(` { "compatibility_version" : "old", "x" : "<x>" } `), []byte(`{"compatibility_version":3}`), []byte(`bad`), []byte{}} {
		var header struct {
			Version string `json:"compatibility_version"`
		}
		legacyErr := json.Unmarshal(raw, &header)
		version, err := DefinitionsVersion(raw)
		if (err == nil) != (legacyErr == nil) || err == nil && version != header.Version {
			t.Fatalf("metadata %s: %q %v != %q %v", raw, version, err, header.Version, legacyErr)
		}
		want, wantErr := json.Marshal(raw)
		for range 2 {
			compact, err := CompactDefinitions(raw)
			if (err == nil) != (wantErr == nil) || err == nil && string(compact) != string(want) {
				t.Fatalf("compact %s: %q %v != %q %v", raw, compact, err, want, wantErr)
			}
			if len(compact) > 0 {
				compact[0] = 'x'
			} // caller cannot poison a cached raw message
		}
	}
}
