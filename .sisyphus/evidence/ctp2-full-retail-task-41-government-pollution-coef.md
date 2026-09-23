# CTP2 Full Retail Task 41: government pollution coefficient

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:1041-1125` (`CityData::CalcPollution`) multiplies both population and production pollution by `g_player[m_owner]->GetPollutionCoef()` before applying building pollution percentages.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/government.cdb:15` defines `PollutionCoef` for governments.
- Retail `govern.txt` includes `PollutionCoef` values such as Democracy `0.8`.

## Implementation
- Runtime governments now include safe public `pollution_percent` derived from `(PollutionCoef - 1) * 100`, preserving reducers and penalties.
- Setting a government upserts a nation `pollution_percent` effect.
- City pollution applies government pollution percent to base population and production pollution before building population/production reducers, matching CTP2 ordering.
- Synthetic rule validation accepts `PollutionPercent` for government fixtures, while existing retail startup remains opt-in.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailCityPollutionAppliesGovernmentCoefficientBeforeBuildingReducers|TestGovernmentRuntimeYieldPercentEffects|TestRetailCityPollutionUsesPopulationProductionAndReducers' -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Difficulty pollution thresholds/ratios, wonder pollution modifiers, pollution happiness, and broader government systems (readiness, support, martial law, sliders) remain incomplete.
