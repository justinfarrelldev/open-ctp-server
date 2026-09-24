# CTP2 Full Retail Task 63: overcrowding level effects

## Source references
- `CityData::GetBuildingOvercrowdingBonus` calls `buildingutil_GetRaiseOvercrowdingLevel` and adds the summed result to city overcrowding before growth-rate calculation.
- `buildingutil_GetRaiseOvercrowdingLevel` sums `RaiseOvercrowdingLevel` across effective buildings.
- Retail `buildings.txt` uses `RaiseOvercrowdingLevel` on aqueduct/arcology-style population infrastructure.

## Implementation
- Retail building projection maps `RaiseOvercrowdingLevel` to an `overcrowding_level` effect.
- City recalculation now sums effective overcrowding-level effects and stores the public city value.
- Owner projections include the derived overcrowding level without exposing hidden state.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestRetailOvercrowdingEffectsAccumulateFromBuildings -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/visibility -count=1`

## Remaining gaps
- The simplified food-growth model does not yet reproduce the full CTP2 city-size growth-rate curve that consumes this overcrowding value.
