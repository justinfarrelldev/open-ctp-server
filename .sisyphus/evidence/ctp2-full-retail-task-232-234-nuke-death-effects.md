# CTP2 Full Retail Tasks 232-234: nuke blast unit casualties and death-happiness marker

## Implementation
- Imported retail `DeathEffectsHappy` unit flags as the safe runtime/canonical `death_effects_happy` effect.
- Plant-nuke success now removes non-acting surface units on the target city tile and adjacent tiles, reporting only a safe `destroyed_units` count in the event payload.
- Units with `death_effects_happy` killed by the nuke blast are registered through the existing peace-movement loss accounting.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestRetailPlantNukeRemovesUnitsInBlastRadiusAndRecordsDeathEffect -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/visibility -count=1`
- `go test ./internal/gameplay/ai -count=1`
