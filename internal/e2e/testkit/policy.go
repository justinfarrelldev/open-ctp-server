package testkit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	innerTimeoutFlag = "-timeout="
	outerTimeout     = "timeout --signal=TERM --kill-after=30s"
)

// ValidateTestCommand enforces the bounded test process policy used by local tasks and CI.
func ValidateTestCommand(command string) error {
	for _, required := range []string{"GOMAXPROCS=4", outerTimeout, "go test", innerTimeoutFlag, "-p=2"} {
		if !strings.Contains(command, required) {
			return fmt.Errorf("test command is missing %q", required)
		}
	}
	return nil
}

// ValidateCoverage rejects a package result below its required percentage.
func ValidateCoverage(packageName string, actual, minimum float64) error {
	if actual < minimum {
		return fmt.Errorf("%s coverage %.1f%% is below required %.1f%%", packageName, actual, minimum)
	}
	return nil
}

// ValidateRepositoryTimeoutPolicy checks every Go test invocation in the supplied files.
func ValidateRepositoryTimeoutPolicy(paths ...string) error {
	for _, path := range paths {
		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		text := string(contents)
		cursor := 0
		for {
			index := strings.Index(text[cursor:], "go test")
			if index < 0 {
				break
			}
			index += cursor
			start := strings.LastIndex(text[:index], "\n") + 1
			endOffset := strings.IndexByte(text[index:], '\n')
			end := len(text)
			if endOffset >= 0 {
				end = index + endOffset
			}
			line := text[start:end]
			if err := ValidateTestCommand(line); err != nil {
				return fmt.Errorf("%s: %w", filepath.Base(path), err)
			}
			cursor = index + len("go test")
		}
	}
	return nil
}

// RepositoryRoot finds the checkout from a package test working directory.
func RepositoryRoot(start string) (string, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "Taskfile.yml")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("Taskfile.yml not found above %s", start)
		}
		current = parent
	}
}
