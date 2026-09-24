# CTP2 Full Retail Task 77: bombard stat preservation

## Source references
- Retail `Units.txt` uses `BombardRange`, `BombRounds`, and `ZBRangeAttack` for ranged/bombard units.
- `unit.cdb` defines bombard range/round fields as gameplay-affecting unit stats.

## Implementation
- Canonical units now store `bombard_range` and `bomb_rounds` alongside ranged attack.
- Runtime unit production copies bombard range/rounds from retail runtime definitions.
- Owner projections expose these owned-unit stats safely for legal bombard/action planning.

## Tests
- `go test ./internal/gameplay/engine -run TestProductionCompletionCopiesRuntimeRangedStats -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- The active bombard order itself still needs full retail execution.
