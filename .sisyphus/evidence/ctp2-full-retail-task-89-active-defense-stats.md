# CTP2 Full Retail Task 89: active-defense stat preservation

## Source references
- Retail `Units.txt`/unit catalog contains `ActiveDefenseRange` for air-defense/counter-bombard style units.
- The value is gameplay-affecting and must survive retail runtime import and unit production.

## Implementation
- Runtime unit definitions now include `active_defense_range`.
- Retail runtime projection maps `ActiveDefenseRange` into runtime unit definitions.
- Produced units and owner unit projections preserve active-defense range.
- State validation rejects negative active-defense range values.

## Tests
- `go test ./internal/gameplay/engine -run TestProductionCompletionCopiesRuntimeRangedStats -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Active-defense interception execution still needs full retail combat semantics.
