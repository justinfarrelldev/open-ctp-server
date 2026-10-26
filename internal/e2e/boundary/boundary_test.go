package boundary

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/testkit"
)

func TestProductionPackagesDoNotImportOrExposeE2ECode(t *testing.T) {
	root, err := testkit.RepositoryRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	productionFiles, err := collectProductionGoFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range productionFiles {
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, imported := range parsed.Imports {
			value, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				t.Fatalf("decode import in %s: %v", path, err)
			}
			if strings.Contains(value, "/internal/e2e/") {
				t.Fatalf("production file %s imports E2E package %q", relative(root, path), value)
			}
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(literal.Value)
			if err != nil {
				return true
			}
			lower := strings.ToLower(value)
			if strings.Contains(lower, "diagnostic") || strings.Contains(lower, "puppet") {
				t.Errorf("production file %s contains E2E route/topic/config string %q", relative(root, path), value)
			}
			return true
		})
		for _, comment := range parsed.Comments {
			if strings.Contains(comment.Text(), "go:build test") {
				t.Errorf("production file %s contains a nonstandard test build tag", relative(root, path))
			}
		}
	}
	if _, err := os.Stat(filepath.Join(root, "internal", "gameplay", "testkit")); !os.IsNotExist(err) {
		t.Fatalf("production gameplay testkit still exists: %v", err)
	}
}

func TestProductionMainRunsAutonomousAIWorker(t *testing.T) {
	root, err := testkit.RepositoryRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(root, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(contents)
	if !strings.Contains(text, `go runtime.AIWorker.Run(context.Background())`) {
		t.Fatal("main.go does not start the autonomous AI worker")
	}
}

func TestProductionAIGatewayRemainsSeparateFromHumanCommandExecution(t *testing.T) {
	root, err := testkit.RepositoryRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "internal", "gameplay", "service", "ai.go")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(contents)
	if !strings.Contains(text, ".ExecuteAICommand(") {
		t.Fatal("production AI gateway no longer traverses ExecuteAICommand")
	}
	if strings.Contains(text, ".ExecuteCommand(") {
		t.Fatal("production AI gateway traverses the authenticated human command method")
	}
}

func TestTaskFourPackagesHaveNoBuildOrRuntimeSwitches(t *testing.T) {
	root, err := testkit.RepositoryRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, directory := range []string{"diagnostic", "puppet", "testkit"} {
		path := filepath.Join(root, "internal", "e2e", directory)
		err := filepath.WalkDir(path, func(file string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(file, ".go") {
				return nil
			}
			contents, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			text := string(contents)
			for _, forbidden := range []string{"//go:build", "os.Getenv(", "os.LookupEnv("} {
				if strings.Contains(text, forbidden) {
					t.Errorf("Task 4 E2E file %s contains forbidden runtime/build switch %q", relative(root, file), forbidden)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func collectProductionGoFiles(root string) ([]string, error) {
	paths := []string{filepath.Join(root, "main.go")}
	internal := filepath.Join(root, "internal")
	err := filepath.WalkDir(internal, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path == filepath.Join(internal, "e2e") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

func relative(root, path string) string {
	value, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(value)
}
