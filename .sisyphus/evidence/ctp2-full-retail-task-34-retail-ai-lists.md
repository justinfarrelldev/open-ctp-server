# CTP2 full-retail parity task 34 retail AI list projection evidence

Scope: import released retail AI list files, project them through runtime/public rules safely, and make AI production use projected retail list order.

Implemented:
- `LoadRetailRulesFromDirectory` now also loads `default/aidata` list files adjacent to `gamedata`:
  - `UnitBuildLists.txt`
  - `BuildingBuildLists.txt`
  - `WonderBuildLists.txt`
  - `AdvanceLists.txt`
  - `ImprovementLists.txt`
- Added typed `RetailAILists` / `RetailAIList` and runtime-safe `RuntimeAI` / `RuntimeAIList` projections.
- Runtime projection preserves released list order and item IDs without exposing hidden state.
- Viewer rules projection now includes cloned runtime AI lists.
- AI city production now honors projected retail building list order before fallback scoring.
- Refreshed 100-turn retail AI transcript after the runtime content/projection change.

Validation commands:
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/visibility -run RulesProjectionIncludesRetailAILists -count=1`
- `go test ./internal/gameplay/ai -run 'PlannerUsesRetailAIBuildingListOrder|PlannerQueuesProjectedRetailYieldBuildingBeforeUnits|PlannerQueuesProjectedRetailMilitaryAfterExpansion' -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `GOMAXPROCS=4 go test ./... -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`
