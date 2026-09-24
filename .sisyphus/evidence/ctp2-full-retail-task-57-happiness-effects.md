# CTP2 Full Retail Task 57: retail happiness effects

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/buildingutil.cpp:123-169` accumulates building happiness increments.
- Retail `buildings.txt` uses `HappyInc` on improvements such as `IMPROVE_ARENA`.
- Retail `Wonder.txt` uses `IncHappinessEmpire` on wonders such as `WONDER_GUTTENBERGS_BIBLE`.

## Implementation
- Runtime projection maps building `HappyInc` and wonder `IncHappinessEmpire` to safe `happiness` effects, including signed values.
- City recalculation now accumulates city/building/wonder/nation `happiness` effects into a public `City.Happiness` field.
- Owner city projections include the derived happiness value without exposing hidden city state to other viewers.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestRetailHappinessEffectsAccumulateFromBuildingsAndWonders -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/visibility -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Full retail happiness accounting, crime, overcrowding, martial law, riots, specialists, and sliders remain incomplete; this task adds the effect plumbing and city projection foundation.
