package main

import (
	"fmt"
	"os"
	"testing"
)

// TestMain is the entry point for running tests in this package. It runs all the tests
// and then checks the code coverage if coverage mode is enabled. If the tests pass
// (i.e., return code is 0) and coverage mode is enabled, it verifies that the code
// coverage is at least 80%. If the coverage is below 80%, it prints a message indicating
// that the tests passed but the coverage failed, and sets the return code to -1 to
// indicate failure. Finally, it exits with the appropriate return code.
//
// Parameters:
// - m: *testing.M, the test manager that provides methods to run tests and benchmarks.
//
// Behavior:
// - Runs all tests using m.Run().
// - If tests pass and coverage mode is enabled, checks the code coverage.
// - If coverage is below 80%, prints a failure message and sets the return code to -1.
// - Exits with the final return code.
func TestMain(m *testing.M) {
	rc := m.Run()

	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < 0.8 {
			fmt.Printf("Tests passed but coverage failed at %.2f%%\n", c*100)
			rc = -1
		}
	}
	os.Exit(rc)
}
