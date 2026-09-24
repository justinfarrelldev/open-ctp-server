# CTP2 Full Retail Task 78: bombard command

## Source references
- Retail units use `ZBRangeAttack`, `BombardRange`, and `BombRounds` for ranged bombardment.
- CTP2 combat has distinct ranged/bombard orders separate from adjacent melee capture.

## Implementation
- Added a `bombard` command and `bombard.resolved` event.
- Owned units with ranged attack, bombard range, bombard rounds, movement, visible target tile, and target compatibility can bombard enemy units at range.
- Bombard uses deterministic combat RNG, applies ranged damage for configured rounds, removes destroyed targets, and spends one movement point without capturing cities.
- Owner legal-action projection now exposes safe bombard actions against visible enemy occupied tiles.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailBombardDamagesVisibleTargetAtRange -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Full retail bombard/counter-bombard/active-defense semantics and target-type masks still need deeper implementation.
