# CTP2 Full Retail Task 235: preserve retail Explodes unit marker

## Implementation
- Imported retail `Explodes` unit flags as the safe runtime/canonical `explodes` effect.
- This preserves the gameplay-affecting marker for future explosion-specific execution without claiming full execution parity yet.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
