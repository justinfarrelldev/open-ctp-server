# CTP2 Full Retail Task 122: plant-nuke chance

## Implementation
- `special_plant_nuke` positive effect values now act as deterministic percent success chances.
- Failed chance rolls spend the unit move and emit `city.nuke_failed` without changing population or pollution.
- Zero-value legacy/synthetic nuke effects remain 100% success for existing fixtures.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailPlantNukeDamagesCityAndHonorsProtection -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail escape/death chances and exact nuclear side effects remain future work.
