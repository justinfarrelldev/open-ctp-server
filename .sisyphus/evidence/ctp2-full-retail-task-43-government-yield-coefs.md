# CTP2 Full Retail Task 43: signed government yield coefficients

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/government.cdb:49-51` defines `FoodCoef`, `GoldCoef`, and `ProductionCoef`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:1202`, `1483`, and `5854` apply government production, food, and gold coefficients to city yields.
- Retail `govern.txt` contains both bonuses and penalties, e.g. Democracy food `1.25` and Anarchy food/production/gold `0.75`.

## Implementation
- Runtime government projection now exposes signed `food_percent`, `production_percent`, and `gold_percent` deltas from retail coefficients.
- Government changes upsert food/production/gold percent effects even when negative.
- City yield percent effects now support signed percentages with retail-style ceil rounding for adjusted yields.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestGovernmentRuntimeYieldPercentEffects|TestSignedYieldPercentEffectsCanReduceCityYield|TestRuntimeProductionEffectsApplyToCompletedBuildingYields' -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Knowledge/science coefficient is still represented by the existing bounded research bonus model, not the full CTP2 slider/science allocation pipeline.
- Government rations, wages, workday, readiness, support, martial law, rush-buy, trade caps, and war discontent remain incomplete.
