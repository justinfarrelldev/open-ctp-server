# Task 051: retail InvisibleForGlobeSat hides units from global radar-only visibility

## Retail source/data checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:1873-1882`
  - At turn-vision refresh, global radar adds visibility only when the viewer has global radar and the unit DB record reports standard vision class.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Units.txt:60` and additional unit records
  - Retail data marks many units with `InvisibleForGlobeSat`.
- `internal/gameplay/rules/retail_runtime.go:585`
  - The importer maps `InvisibleForGlobeSat` into the runtime `invisible_for_globe_sat` effect.

## Implementation notes

- Global radar still reveals every tile and normal foreign units.
- A foreign unit with `EffectInvisibleForGlobeSat` is not projected when the tile is visible only because of global radar.
- Normal sight from a unit/city/improvement still reveals the unit, so this is not a general stealth rule and does not grant hidden-state shortcuts.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-051-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/visibility -run 'TestRetailInvisibleForGlobeSat|TestRetailGlobalRadarRevealsAllTilesAndUnits' -count=1`
