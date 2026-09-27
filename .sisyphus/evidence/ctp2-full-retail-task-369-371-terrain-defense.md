# CTP2 Full Retail Tasks 369-371: terrain defense modifiers

## Source notes
- Retail `terrain.txt` `EnvBase.Defense` is a terrain defense coefficient (for example forest is `0.25`).
- Retail combat applies terrain defensive value to defenders on those tiles.

## Implementation
- Runtime terrain projection now preserves `defense_bonus_percent` from retail `EnvBase.Defense`.
- Combat, bombard defense, active defense, and counter-bombard defense use runtime terrain defense modifiers when runtime terrain data is present, retaining the legacy forest fallback for synthetic tests.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetail(TerrainDefenseBonusModifiesDefenders|CityDefenseBuildingEffectsModifyDefenders|TerrainImprovementDefenseBonusModifiesDefenders)' -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
- `go test ./internal/gameplay/ai -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`
- `GOMAXPROCS=4 go test ./... -count=1`
- `git diff --check`
