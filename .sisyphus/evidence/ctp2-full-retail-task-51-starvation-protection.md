# CTP2 Full Retail Task 51: building starvation protection

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/building.cdb:48` defines `StarvationProtection`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/buildingutil.cpp:447-456` sums building starvation protection.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:1590-1619` delays starvation population loss while protection turns remain.
- Retail `buildings.txt` gives `IMPROVE_GRANARY` `StarvationProtection 2` and `IMPROVE_FOOD_SILO` `StarvationProtection 5`.

## Implementation
- Runtime building projection maps retail `StarvationProtection` to a safe `starvation_protection` effect.
- City food advancement now tracks consecutive starvation turns and delays population loss up to the summed city protection effects, resetting the counter once the city is no longer starving.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetailStarvationProtectionDelaysPopulationLoss|TestRetailCityFoodSurplusGrowsPopulation|TestRetailBuildingMaxPopulationCapsFoodGrowth' -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Full CTP2 growth-rate/partial-population model, base starvation protection constants, warning events, and hunger/happiness side effects remain incomplete.
