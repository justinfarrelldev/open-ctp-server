package world

import (
	"encoding/json"
	"fmt"
	"testing"
)

// CanonicalJSON returns the stable JSON encoding used by the world harness.
// It deliberately avoids indentation so identical values compare byte-for-byte.
func CanonicalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

// CanonicalString returns the stable JSON encoding as a string.
func CanonicalString(v any) (string, error) {
	bytes, err := CanonicalJSON(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// MustCanonicalJSON is a test helper for concise serialization assertions.
func MustCanonicalJSON(t testing.TB, v any) []byte {
	t.Helper()

	bytes, err := CanonicalJSON(v)
	if err != nil {
		t.Fatalf("CanonicalJSON() error = %v", err)
	}
	return bytes
}

// MustCanonicalString is a test helper for concise serialization assertions.
func MustCanonicalString(t testing.TB, v any) string {
	t.Helper()

	text, err := CanonicalString(v)
	if err != nil {
		t.Fatalf("CanonicalString() error = %v", err)
	}
	return text
}

// AssertCanonicalJSONEqual compares two values by their canonical JSON payloads.
func AssertCanonicalJSONEqual(t testing.TB, got, want any) {
	t.Helper()

	gotBytes := MustCanonicalJSON(t, got)
	wantBytes := MustCanonicalJSON(t, want)
	if string(gotBytes) != string(wantBytes) {
		t.Fatalf("canonical JSON mismatch\ngot:  %s\nwant: %s", gotBytes, wantBytes)
	}
}

// CanonicalSnapshotKey produces a compact, deterministic key for snapshot comparisons.
func CanonicalSnapshotKey(v any) (string, error) {
	text, err := CanonicalString(v)
	if err != nil {
		return "", fmt.Errorf("canonical snapshot key: %w", err)
	}
	return text, nil
}
