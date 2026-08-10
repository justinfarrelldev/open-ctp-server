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
	CompatibilityVersion string `json:"compatibility_version"`
	ContentHash          string `json:"content_hash"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("import-rules", flag.ContinueOnError)
	flags.SetOutput(stderr)
	dataDir := flags.String("data-dir", "", "user-owned CTP2 gamedata directory")
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

	ruleset, err := rules.Import(rules.ImportConfig{Directory: *dataDir})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(rulesetIdentity{
		CompatibilityVersion: ruleset.CompatibilityVersion,
		ContentHash:          ruleset.ContentHash,
	}); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
