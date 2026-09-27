# CTP2 Full Retail Task 227: single-use attackers are consumed after combat

## Implementation
- Surviving units with `single_use` are removed after resolving a normal combat command.
- This complements bombard consumption so missile-style retail units cannot persist after use.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailSingleUse(AttackerIsConsumedAfterCombat|BombarderIsConsumedAfterBombardment)' -count=1`
- `go test ./internal/gameplay/engine -count=1`
