# CTP2 Full Retail Task 52: air-unit refuel and airport projection

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:5658-5689` refuels max-fuel units when transported or when located in a city, and also at owned airfields.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/buildingutil.cpp:288-294` and `CityData.cpp:3033-3036` expose city airport status from building records.
- Retail `buildings.txt` marks `IMPROVE_AIRPORT` with `Airport`.

## Implementation
- Runtime building projection now exposes a safe `airport` effect for retail airport buildings.
- Turn advancement now refuels max-fuel units to `MaxFuel` when they are transported or on a friendly city tile, while preserving existing deterministic movement refresh behavior.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestAirUnitsRefuelAtFriendlyCitiesOnAdvanceRound|TestUnitMovementSpendsFuelAndRejectsEmptyFuel' -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Airfields as tile improvements, airport/airfield airlift behavior, and exact retail movement-zeroing side effects on refuel remain incomplete.
