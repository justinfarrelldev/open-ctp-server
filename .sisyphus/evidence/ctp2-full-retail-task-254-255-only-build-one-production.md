# CTP2 Full Retail Tasks 254-255: unit production queue repeat and OnlyBuildOne

## Retail source
- `ctp2_source_code/ctp2_code/gs/gameobj/bldque.cpp` keeps a single queued normal unit after completion for continuous production, but removes completed units when another queue item follows or the unit record has `OnlyBuildOne`.

## Implementation
- Runtime unit production now preserves a sole non-`OnlyBuildOne` unit queue item after completion, producing repeated units across turns.
- Runtime unit production consumes the queue item for units marked `only_build_one`.
- Legacy/nil-runtime production behavior is unchanged.

## Tests
- `go test ./internal/gameplay/engine -run TestRuntimeUnitProductionRepeatsUnlessOnlyBuildOne -count=1`
- `go test ./internal/gameplay/engine -count=1`
