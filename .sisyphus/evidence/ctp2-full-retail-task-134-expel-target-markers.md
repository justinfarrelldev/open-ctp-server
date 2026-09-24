# CTP2 Full Retail Task 134: expel/sue target marker import and expel gating

## Implementation
- Retail unit flags `CanBeSued` and `CanBeExpelled` now import into runtime starter effects.
- Engine state validation accepts the new target marker effects.
- `CommandExpel` now requires the target unit to have the `can_be_expelled` marker, matching retail unit targeting restrictions.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailExpel(MovesAdjacentEnemyUnit|RequiresTargetCanBeExpelled)' -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeIncludesCompleteReleasedCatalog -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`

## Remaining gaps
- Lawsuit commands targeting `CanBeSued` units remain future work.
