package freeze

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngineLock(t *testing.T) {
	root := repoRoot(t)
	lock, err := Load(filepath.Join(root, "internal/e2e/freeze/engine-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(root, lock); err != nil {
		t.Fatal(err)
	}
	if len(lock.Entries) != 51 {
		t.Fatalf("manifest entry count = %d, want 51", len(lock.Entries))
	}
	if len(lock.CampaignWriteAllowlist) == 0 {
		t.Fatal("campaign write allowlist is empty")
	}
}

func TestLockRejectsNegativeCasesInTemporaryRoots(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(string, *Lock) error
		want   string
	}{
		{"changed", func(root string, l *Lock) error {
			return os.WriteFile(filepath.Join(root, filepath.FromSlash(l.Entries[0].Path)), []byte("changed"), 0644)
		}, "hash mismatch"},
		{"added", func(root string, l *Lock) error {
			return os.WriteFile(filepath.Join(root, "internal/gameplay/added.go"), []byte("package added\n"), 0644)
		}, "undeclared production file"},
		{"deleted", func(root string, l *Lock) error {
			return os.Remove(filepath.Join(root, filepath.FromSlash(l.Entries[0].Path)))
		}, "missing file"},
		{"duplicate", func(root string, l *Lock) error { l.Entries = append(l.Entries, l.Entries[0]); return nil }, "duplicate entry"},
		{"out-of-scope", func(root string, l *Lock) error { l.Entries[0].Path = "README.md"; return nil }, "out of scope"},
		{"malformed", func(root string, l *Lock) error { l.Entries[0].Path = "../main.go"; return nil }, "invalid canonical relative path"},
		{"forbidden-import", func(root string, l *Lock) error {
			p := filepath.Join(root, filepath.FromSlash("internal/runtime/runtime.go"))
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if err := os.WriteFile(p, append([]byte("package runtime\n\nimport _ \"example.com/internal/e2e/forbidden\"\n"), b[len([]byte("package runtime\n")):]...), 0644); err != nil {
				return err
			}
			for i := range l.Entries {
				if l.Entries[i].Path == "internal/runtime/runtime.go" {
					l.Entries[i].SHA256, _ = digest(p)
				}
			}
			return nil
		}, "imports E2E package"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, l := tempTree(t)
			if err := tc.mutate(root, &l); err != nil {
				t.Fatal(err)
			}
			err := Validate(root, l)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "../../.."))
}
func tempTree(t *testing.T) (string, Lock) {
	t.Helper()
	source := repoRoot(t)
	lock, err := Load(filepath.Join(source, "internal/e2e/freeze/engine-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, e := range lock.Entries {
		src := filepath.Join(source, filepath.FromSlash(e.Path))
		dst := filepath.Join(root, filepath.FromSlash(e.Path))
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, b, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root, lock
}
func TestRequiredPathsCannotBeOmitted(t *testing.T) {
	t.Run("deleted and manifest entry removed", func(t *testing.T) {
		root, lock := tempTree(t)
		const required = "internal/world/config.go"
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(required))); err != nil {
			t.Fatal(err)
		}
		filtered := lock.Entries[:0]
		for _, entry := range lock.Entries {
			if entry.Path != required {
				filtered = append(filtered, entry)
			}
		}
		lock.Entries = filtered
		err := Validate(root, lock)
		if err == nil || !strings.Contains(err.Error(), `required production path "`+required+`": missing file`) {
			t.Fatalf("error = %v, want missing required path %q", err, required)
		}
	})
}

func TestSymlinkCannotSatisfyRequiredManifestPath(t *testing.T) {
	root, lock := tempTree(t)
	const required = "internal/world/config.go"
	file := filepath.Join(root, filepath.FromSlash(required))
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "main.go"), file); err != nil {
		t.Fatal(err)
	}
	err := Validate(root, lock)
	if err == nil || !strings.Contains(err.Error(), `manifest path "`+required+`": not a regular file`) {
		t.Fatalf("error = %v, want symlink rejection for %q", err, required)
	}
}

func TestManifestPathsUseCleanSlashSyntax(t *testing.T) {
	for _, value := range []string{"internal//world/config.go", "internal/./world/config.go", "internal/world/../world/config.go", "\\internal\\world\\config.go"} {
		if err := validatePath(value); err == nil || !strings.Contains(err.Error(), "invalid canonical relative path") {
			t.Fatalf("validatePath(%q) = %v, want canonical-path error", value, err)
		}
	}
	if err := validatePath("internal/world/config.go"); err != nil {
		t.Fatalf("valid path rejected: %v", err)
	}
}

func TestCampaignWriteAllowlistRejectsTampering(t *testing.T) {
	cases := []struct {
		name   string
		mutate func([]string) []string
		want   string
	}{
		{"empty", func([]string) []string { return []string{} }, "length mismatch"},
		{"missing", func(v []string) []string { return v[:len(v)-1] }, "length mismatch"},
		{"extra", func(v []string) []string { return append(v, "extra") }, "length mismatch"},
		{"reordered", func(v []string) []string { v[0], v[1] = v[1], v[0]; return v }, "mismatch at index 0"},
		{"duplicate", func(v []string) []string { v[1] = v[0]; return v }, "mismatch at index 1"},
		{"changed", func(v []string) []string { v[0] = "changed"; return v }, "mismatch at index 0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, lock := tempTree(t)
			lock.CampaignWriteAllowlist = tc.mutate(append([]string(nil), lock.CampaignWriteAllowlist...))
			err := Validate(root, lock)
			if err == nil || !strings.Contains(err.Error(), "campaign_write_allowlist") || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want campaign_write_allowlist %q", err, tc.want)
			}
		})
	}
}

func TestManifestJSONIsStable(t *testing.T) {
	root := repoRoot(t)
	b, err := os.ReadFile(filepath.Join(root, "internal/e2e/freeze/engine-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	var l Lock
	if err := json.Unmarshal(b, &l); err != nil {
		t.Fatal(err)
	}
	if l.Definition == "" {
		t.Fatal(fmt.Sprint("missing definition"))
	}
}
