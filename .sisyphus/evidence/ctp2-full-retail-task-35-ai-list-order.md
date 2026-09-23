# CTP2 full-retail parity task 35 AI list order evidence

Scope: extend projected retail AI-list usage beyond building production.

Implemented:
- Research selection now honors projected retail advance-list order before cost fallback.
- Unit production now honors projected retail unit-list order before strength/scoring fallback.
- Expansion settler production also honors projected unit-list order for settler-capable units.
- Existing building-list ordering from task 34 remains covered.

Validation commands:
- `go test ./internal/gameplay/ai -run 'PlannerUsesRetailAI(Advance|Unit)ListOrder|PlannerUsesRetailAIBuildingListOrder|PlannerQueuesProjectedRetailLikeUnitProduction|PlannerQueuesProjectedRetailSettlerForExpansion' -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `GOMAXPROCS=4 go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`
