package rules

import "encoding/json"

var definitionVersions = newRulesCache[string](rulesCacheEntries, rulesCacheInputBytes, func(value string) string { return value })

// DefinitionsVersion reads metadata only, not permission to execute a ruleset.
// Every execution still verifies the claimed hash and supported compiled model.
// Exact input bytes key this bounded cache so changed/invalid definitions cannot
// borrow the compatibility decision of a previously valid pin.
func DefinitionsVersion(definitions []byte) (string, error) {
	return definitionVersions.get(definitions, "", func() (string, error) {
		var header struct {
			Version string `json:"compatibility_version"`
		}
		err := json.Unmarshal(definitions, &header)
		return header.Version, err
	})
}

// CompactDefinitions preserves RawMessage's JSON key order, escaping and null
// semantics exactly. CanonicalDefinitions sorts keys and is NOT interchangeable
// here: execution clones must retain their historical canonical storage bytes.
// Returned bytes are owned; the shared cache also bounds this separate namespace.
func CompactDefinitions(definitions json.RawMessage) ([]byte, error) {
	return canonicalRulesCache.get(definitions, "raw-message", func() ([]byte, error) { return json.Marshal(definitions) })
}
