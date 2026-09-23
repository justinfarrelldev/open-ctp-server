// import-rules imports a caller-selected CTP2 data directory and prints its ruleset identity.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

type rulesetIdentity struct {
	CompatibilityVersion string         `json:"compatibility_version"`
	ContentHash          string         `json:"content_hash"`
	RulesetType          string         `json:"ruleset_type,omitempty"`
	RuntimeCounts        *runtimeCounts `json:"runtime_counts,omitempty"`
}

type runtimeCounts struct {
	Units               int `json:"units"`
	Advances            int `json:"advances"`
	Production          int `json:"production"`
	Terrains            int `json:"terrains"`
	TerrainImprovements int `json:"terrain_improvements"`
	Governments         int `json:"governments"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("import-rules", flag.ContinueOnError)
	flags.SetOutput(stderr)
	dataDir := flags.String("data-dir", "", "user-owned CTP2 gamedata directory")
	retail := flags.Bool("retail", false, "import the released-game retail gamedata catalog instead of the bounded synthetic format")
	format := flags.String("format", "identity", "output format: identity or definitions")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "unexpected positional arguments")
		return 2
	}
	if *dataDir == "" {
		fmt.Fprintln(stderr, "--data-dir is required")
		return 2
	}
	if *format != "identity" && *format != "definitions" {
		fmt.Fprintln(stderr, "--format must be identity or definitions")
		return 2
	}

	if *retail {
		return runRetailImport(*dataDir, *format, stdout, stderr)
	}
	return runSyntheticImport(*dataDir, *format, stdout, stderr)
}

func runSyntheticImport(dataDir, format string, stdout, stderr io.Writer) int {
	ruleset, err := rules.Import(rules.ImportConfig{Directory: dataDir})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if format == "definitions" {
		if _, err := stdout.Write(append(ruleset.CanonicalJSON, '\n')); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}
	return writeIdentity(stdout, stderr, rulesetIdentity{CompatibilityVersion: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash})
}

func runRetailImport(dataDir, format string, stdout, stderr io.Writer) int {
	db, err := rules.LoadRetailRulesFromDirectory(dataDir)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	runtime := rules.ProjectRetailRuntime(db)
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if _, err := rules.CompileRuntime(definitions, contentHash); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if format == "definitions" {
		if _, err := stdout.Write(append(definitions, '\n')); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}
	return writeIdentity(stdout, stderr, rulesetIdentity{
		CompatibilityVersion: runtime.CompatibilityVersion,
		ContentHash:          contentHash,
		RulesetType:          "retail-runtime-projection",
		RuntimeCounts: &runtimeCounts{
			Units:               len(runtime.Units),
			Advances:            len(runtime.Advances),
			Production:          len(runtime.Production),
			Terrains:            len(runtime.Movement.Terrains),
			TerrainImprovements: len(runtime.TerrainImprovements),
			Governments:         len(runtime.Governments),
		},
	})
}

func writeIdentity(stdout, stderr io.Writer, identity rulesetIdentity) int {
	if err := json.NewEncoder(stdout).Encode(identity); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
