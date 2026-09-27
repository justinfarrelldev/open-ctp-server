# CTP2 Full Retail Tasks 207-209: runtime restart victory validation

## Implementation
- Runtime projection restart/compile now validates projected victory evaluator IDs against the supported manifest.
- Duplicate projected victory evaluators are rejected instead of silently changing terminal behavior after restart.
- Retail projected runtime canonical definitions now round-trip through `CompileRuntime` while preserving the retail 2300 AD final turn and enabled natural victory evaluators.

## Tests
- `go test ./internal/gameplay/rules -run 'TestCompileRuntimeRejects(UnknownProjectionVictoryEvaluator|DuplicateProjectionVictoryEvaluator)|TestRetailRuntimeProjectionRoundTripsTerminalCalendar' -count=1`
- `go test ./internal/gameplay/rules -count=1`
