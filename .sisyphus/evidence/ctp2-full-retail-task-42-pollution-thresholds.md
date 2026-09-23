# CTP2 Full Retail Task 42: pollution thresholds and ratios

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:1041-1125` subtracts difficulty `PollutionStartPopulationLevel` / `PollutionStartProductionLevel`, clamps negative buckets to zero, then multiplies by difficulty `PollutionPopulationRatio` / `PollutionProductionRatio` before government and building modifiers.
- Retail `DiffDB.txt` defines `POLLUTION_START_POPULATION_LEVEL 16`, `POLLUTION_POPULATION_RATIO 15`, `POLLUTION_START_PRODUCTION_LEVEL 300`, and `POLLUTION_PRODUCTION_RATIO 1` for the shipped difficulties.

## Implementation
- Retail catalog loading now reads the first `DiffDB.txt` difficulty pollution model and preserves it in `RetailRules.PollutionModel`.
- Runtime projection exposes a safe `pollution_model` only when a model is present, avoiding synthetic rules hash/projection drift.
- City pollution now applies population/production thresholds and ratios before government coefficients and building reducers.
- Viewer rules projections clone the optional pollution model as public rules metadata.

## Tests
- `go test ./internal/gameplay/rules -run 'TestLoadRetailRulesFromDirectoryIncludesReleasedCatalog|TestProjectRetailRuntimeIncludesCompleteReleasedCatalog' -count=1`
- `go test ./internal/gameplay/engine -run TestRetailCityPollutionUsesDifficultyThresholdsAndRatios -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Runtime still does not expose selectable difficulty; the bounded retail projection uses the first shipped difficulty model, whose pollution values match the other shipped difficulty records for these fields.
- Pollution happiness/local warning effects and additional pollution sources remain incomplete.
