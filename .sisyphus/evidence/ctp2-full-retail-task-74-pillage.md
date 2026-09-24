# CTP2 Full Retail Task 74: pillage improvements

## Source references
- `ArmyData::CanPillage` checks for tile improvements and a non-air unit with `CanPillage` that can perform a special action.
- `ArmyData::Pillage` spends the special action and queues `GEV_CutImprovements` for the army tile.
- Retail `Units.txt` marks pillaging-capable units with `CanPillage`.

## Implementation
- Retail unit projection maps `CanPillage` to a runtime `can_pillage` effect.
- Added a `pillage` command that requires an owned, untransported unit with `can_pillage` on a non-owned tile containing completed improvements.
- Successful pillage removes completed improvement overlays on the unit tile, preserves pending work, and spends one unit movement point.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailPillageRemovesCompletedEnemyTileImprovements -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/ai ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Full retail diplomacy/war confirmation, AI pillage planning, and exact `CutImprovements` selection semantics still need implementation.
