# CTP2 Full Retail Task 44: research percentage coefficients

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:5791-5804` splits trade into science, applies building science percent, wonder knowledge percent, then multiplies by owner government `KnowledgeCoef`.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/government.cdb:14` defines `KnowledgeCoef`.
- Retail building `SciencePercent` and wonder `IncKnowledgePercent` fields represent percentage modifiers, not flat research points.

## Implementation
- Added `research_percent` effects for retail building science, wonder knowledge, and government knowledge coefficients.
- Research rate now applies city-level research percentages to city trade-derived research, then applies nation/government research percentage to the city subtotal.
- Existing flat `research` effects remain supported for synthetic fixtures and bounded non-retail tests.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetailResearchRateUsesCitySciencePercentEffect|TestRetailResearchRateUsesGovernmentKnowledgeCoefficient|TestGovernmentRuntimeYieldPercentEffects|TestRetailResearchRateUsesCityTradeYield|TestResearchRateSkipsObsoleteRetailBuildingScience' -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Full science tax sliders and specialists are still not modeled; current bounded research assumes city gold/trade is the available science base.
