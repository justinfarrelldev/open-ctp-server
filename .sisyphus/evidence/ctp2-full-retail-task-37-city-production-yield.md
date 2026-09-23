# CTP2 full-retail parity task 37 city production uses full city yield evidence

Scope: production accumulation should use the city's existing economy (worked tiles, completed buildings, wonders, government/nation effects), not only raw city-center terrain production.

Implemented:
- `completeProduction` now accumulates `cityYieldWithRuntime(...).Production` for the existing city state before completing the queued item.
- Newly completed buildings/wonders do not affect their own completion turn because the yield is computed before queue removal/completion.
- Existing direct terrain-production behavior remains covered.

Validation commands:
- `go test ./internal/gameplay/engine -run 'TestRetailCityProductionUsesExistingCityYieldEffects|TestRetailRuntimeTerrainYieldsAccumulateProduction|TestRuntimeProductionEffectsApplyToCompletedBuildingYields' -count=1`
- `GOMAXPROCS=4 go test ./... -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`
