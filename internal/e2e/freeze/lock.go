package freeze

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Entry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type Lock struct {
	Definition             string   `json:"definition"`
	Entries                []Entry  `json:"entries"`
	CampaignWriteAllowlist []string `json:"campaign_write_allowlist"`
}

const definition = "production file = regular .go or declared migration file; excludes _test.go, fixtures, testdata, E2E, scripts, generated evidence, and all other files"

var campaignWriteAllowlist = []string{
	"internal/e2e/** except internal/e2e/freeze/**",
	"synthetic fixtures and scripts/catalog data",
	".github/workflows/**",
	"Taskfile.yml",
	"README.md",
	".sisyphus/evidence/**",
}

func Load(path string) (Lock, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Lock{}, err
	}
	var l Lock
	if err := json.Unmarshal(b, &l); err != nil {
		return Lock{}, fmt.Errorf("lock: invalid JSON: %w", err)
	}
	return l, nil
}

func Validate(root string, l Lock) error {
	if l.Definition != definition {
		return fmt.Errorf("manifest definition: expected %q", definition)
	}
	if err := validateCampaignWriteAllowlist(l.CampaignWriteAllowlist); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, e := range l.Entries {
		if seen[e.Path] {
			return fmt.Errorf("manifest path %q: duplicate entry", e.Path)
		}
		seen[e.Path] = true
	}
	if !sort.SliceIsSorted(l.Entries, func(i, j int) bool { return l.Entries[i].Path < l.Entries[j].Path }) {
		return fmt.Errorf("manifest path %q: entries are not sorted", firstPath(l.Entries))
	}
	for _, e := range l.Entries {
		if err := validatePath(e.Path); err != nil {
			return err
		}
		if !allowed(e.Path) {
			return fmt.Errorf("manifest path %q: out of scope", e.Path)
		}
		if len(e.SHA256) != sha256.Size*2 || strings.ToLower(e.SHA256) != e.SHA256 {
			return fmt.Errorf("manifest path %q: SHA-256 must be lowercase hexadecimal", e.Path)
		}
		if _, err := hex.DecodeString(e.SHA256); err != nil {
			return fmt.Errorf("manifest path %q: invalid SHA-256: %w", e.Path, err)
		}
		info, err := os.Lstat(filepath.Join(root, filepath.FromSlash(e.Path)))
		if err != nil {
			return fmt.Errorf("manifest path %q: missing file: %w", e.Path, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("manifest path %q: not a regular file", e.Path)
		}
		digest, err := digest(filepath.Join(root, filepath.FromSlash(e.Path)))
		if err != nil {
			return fmt.Errorf("manifest path %q: hash: %w", e.Path, err)
		}
		if digest != e.SHA256 {
			return fmt.Errorf("manifest path %q: hash mismatch: manifest %s, actual %s", e.Path, e.SHA256, digest)
		}
	}
	actual, err := Enumerate(root)
	if err != nil {
		return err
	}
	for _, p := range actual {
		if !seen[p] {
			return fmt.Errorf("manifest path %q: undeclared production file", p)
		}
	}
	if len(actual) != len(l.Entries) {
		return fmt.Errorf("manifest: production file count mismatch: manifest %d, actual %d", len(l.Entries), len(actual))
	}
	return ScanProductionImports(root)
}

func validateCampaignWriteAllowlist(actual []string) error {
	if len(actual) != len(campaignWriteAllowlist) {
		return fmt.Errorf("campaign_write_allowlist: length mismatch: got %d, want %d", len(actual), len(campaignWriteAllowlist))
	}
	for i, want := range campaignWriteAllowlist {
		if actual[i] != want {
			return fmt.Errorf("campaign_write_allowlist: mismatch at index %d: got %q, want %q", i, actual[i], want)
		}
	}
	return nil
}

func Enumerate(root string) ([]string, error) {
	set := map[string]bool{}
	for _, locked := range []string{"internal/runtime", "internal/gameplay"} {
		err := filepath.WalkDir(filepath.Join(root, filepath.FromSlash(locked)), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("required production path %q: symlink is not allowed", filepath.ToSlash(rel(root, path)))
			}
			if filepath.Ext(path) == ".go" && !strings.HasSuffix(d.Name(), "_test.go") {
				set[filepath.ToSlash(rel(root, path))] = true
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("enumerate %s: %w", locked, err)
		}
	}
	for _, p := range allowedPaths {
		info, err := os.Lstat(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("required production path %q: missing file", p)
			}
			return nil, fmt.Errorf("required production path %q: cannot inspect file: %w", p, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("required production path %q: not a regular file", p)
		}
		set[p] = true
	}
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

var allowedPaths = []string{"internal/world/config.go", "internal/world/contracts.go", "internal/world/generator.go", "internal/world/metadata.go", "internal/world/models.go", "internal/world/recalculation.go", "internal/world/rng.go", "internal/world/ruins.go", "internal/world/serialization.go", "internal/world/snapshot.go", "internal/world/store.go", "internal/game/create_game.go", "internal/game/endpoints.go", "internal/game/game_handler.go", "internal/game/helpers.go", "internal/game/models.go", "internal/game/player_endpoints.go", "internal/game/start_gameplay.go", "internal/game/world_contract_stubs.go", "internal/game/world_endpoints.go", "internal/game/world_ratings_pollution_endpoints.go", "main.go", "supabase/migrations/20260719120000_gameplay_persistence_task2.sql", "supabase/migrations/20260719130000_gameplay_event_ids_task17.sql", "supabase/migrations/20260719140000_game_setup_dimensions_task17.sql", "supabase/migrations/20260810160000_gameplay_ai_exploration_history_task17.sql"}

func ScanProductionImports(root string) error {
	files, err := Enumerate(root)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	for _, p := range files {
		if !strings.HasSuffix(p, ".go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(root, filepath.FromSlash(p)), nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("production file %q: parse imports: %w", p, err)
		}
		for _, imp := range f.Imports {
			v, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return fmt.Errorf("production file %q: invalid import literal %q: %w", p, imp.Path.Value, err)
			}
			if strings.Contains(v, "/internal/e2e/") || strings.HasSuffix(v, "/internal/e2e") {
				return fmt.Errorf("production file %q imports E2E package %q", p, v)
			}
		}
	}
	return nil
}

func allowed(p string) bool {
	for _, x := range allowedPaths {
		if p == x {
			return true
		}
	}
	return strings.HasPrefix(p, "internal/runtime/") || strings.HasPrefix(p, "internal/gameplay/")
}
func validatePath(p string) error {
	if p == "" || filepath.IsAbs(p) || strings.Contains(p, "\\") || filepath.ToSlash(p) != p || path.Clean(p) != p || p == "." || strings.HasPrefix(p, "../") {
		return fmt.Errorf("manifest path %q: invalid canonical relative path", p)
	}
	return nil
}
func digest(p string) (string, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}
func rel(root, p string) string { r, _ := filepath.Rel(root, p); return r }
func firstPath(es []Entry) string {
	if len(es) == 0 {
		return "<empty>"
	}
	return es[0].Path
}
