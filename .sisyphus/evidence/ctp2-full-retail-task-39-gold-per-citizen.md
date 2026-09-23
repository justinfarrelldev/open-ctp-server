# CTP2 full-retail parity task 39 GoldPerCitizen building effect evidence

Scope: execute the retail building `GoldPerCitizen` field as a city economy effect.

Implemented:
- Added `yield_gold_per_population` effect kind.
- Retail runtime projection maps building `GoldPerCitizen` to `yield_gold_per_population`.
- City yield execution adds `Value * positive population` to gold/trade yield.
- Runtime production/API projection carries the effect as safe public metadata like other production effects.
- Refreshed the 100-turn retail AI transcript after the runtime projection changed.

Validation commands:
- `go test ./internal/gameplay/engine -run TestRuntimeGoldPerCitizenEffectUsesPopulation -count=1`
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `GOMAXPROCS=4 go test ./... -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`
