# CTP2 Full Retail Task 47: building max-population growth caps

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:1519-1528` exposes building-driven max population increases via `buildingutil_GetRaiseMaxPopulation`.
- Retail `buildings.txt` uses `RaiseMaxPopulation 14` on population-infrastructure buildings such as `IMPROVE_AQUEDUCT`.

## Implementation
- Runtime building projection maps `RaiseMaxPopulation` to safe `max_population` effects.
- City food growth now honors the highest `max_population` effect and will not grow beyond that cap.
- Existing uncapped/synthetic growth behavior is preserved when no max-population effect is present.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestRetailBuildingMaxPopulationCapsFoodGrowth -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- The full CTP2 city-size database/growth-rate pipeline and overcrowding happiness effects are still incomplete.
