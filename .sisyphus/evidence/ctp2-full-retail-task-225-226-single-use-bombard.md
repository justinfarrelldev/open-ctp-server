# CTP2 Full Retail Tasks 225-226: single-use unit preservation and bombard consumption

## Implementation
- Imported retail `SingleUse` unit flags into runtime starter effects as `single_use`.
- Added canonical `single_use` effect support.
- Bombard units with `single_use` are consumed after resolving bombardment, matching retail missile-style behavior.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailSingleUseBombarderIsConsumedAfterBombardment -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
