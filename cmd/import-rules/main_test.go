package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRequiresExplicitDataDirectory(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := run(nil, &stdout, &stderr); got != 2 {
		t.Fatalf("run() exit code = %d, want 2", got)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "--data-dir is required") {
		t.Fatalf("stdout=%q stderr=%q, want required flag error only", stdout.String(), stderr.String())
	}
}

func TestRunRejectsUnknownDirectory(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := run([]string{"--data-dir", filepath.Join(t.TempDir(), "missing")}, &stdout, &stderr); got != 1 {
		t.Fatalf("run() exit code = %d, want 1", got)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "invalid_directory") {
		t.Fatalf("stdout=%q stderr=%q, want source-located importer error", stdout.String(), stderr.String())
	}
}

func TestRunEmitsDeterministicRulesetIdentity(t *testing.T) {
	dataDir, err := filepath.Abs(filepath.Join("..", "..", "internal", "gameplay", "rules", "testdata", "synthetic-ctp2"))
	if err != nil {
		t.Fatalf("Abs() error = %v", err)
	}
	first := execute(t, dataDir)
	second := execute(t, dataDir)
	if first != second {
		t.Fatalf("repeat output differs:\nfirst=%q\nsecond=%q", first, second)
	}
	var identity struct {
		CompatibilityVersion string `json:"compatibility_version"`
		ContentHash          string `json:"content_hash"`
	}
	if err := json.Unmarshal([]byte(first), &identity); err != nil {
		t.Fatalf("JSON output error = %v output=%q", err, first)
	}
	if identity.CompatibilityVersion == "" || len(identity.ContentHash) != 64 || strings.Contains(first, "UNIT_SYNTHETIC") {
		t.Fatalf("identity = %+v output=%q, want version/hash only", identity, first)
	}
}

func TestRunRejectsUnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := run([]string{"--installation-dir", "anywhere"}, &stdout, &stderr); got != 2 {
		t.Fatalf("run() exit code = %d, want 2", got)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("stdout=%q stderr=%q, want unknown flag error", stdout.String(), stderr.String())
	}
}

func execute(t *testing.T, dataDir string) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if got := run([]string{"--data-dir", dataDir}, &stdout, &stderr); got != 0 {
		t.Fatalf("run() exit code = %d stderr=%q", got, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
	return stdout.String()
}
