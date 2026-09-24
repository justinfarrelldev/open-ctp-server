# CTP2 Full Retail Task 58: retail crime effects

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/buildingutil.cpp:172-181` accumulates `LowerCrime` from built improvements.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/wonderutil.cpp:130-133` exposes wonder `DecCrimePercent`.
- Retail `buildings.txt` uses `LowerCrime` on buildings such as `IMPROVE_COURTHOUSE`; retail `Wonder.txt` uses `DecCrimePercent` on `WONDER_ZERO_CRIME_BILL`.

## Implementation
- Runtime projection maps building `LowerCrime` to signed `crime_percent` effects.
- Runtime projection maps wonder `DecCrimePercent` to negative `crime_percent` effects.
- City recalculation now derives and projects a public `CrimePercent` aggregate for owned cities.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetailCrimeEffectsAccumulateFromBuildingsAndWonders|TestRetailHappinessEffectsAccumulateFromBuildingsAndWonders' -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/visibility -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Crime does not yet reduce food/production/gold/science or interact with happiness/riots exactly as retail does.
