# CTP2 Full Retail Task 59: retail prerequisite buildings

## Source references
- Retail `buildings.txt` uses `PrerequisiteBuilding` to gate improvements such as `IMPROVE_BANK` behind `IMPROVE_BAZAAR`.
- CTP2 build-queue legality requires both advance and building prerequisites before city production can be queued.

## Implementation
- Runtime production definitions now include safe public `required_buildings` metadata.
- Retail building projection maps `PrerequisiteBuilding` into runtime production requirements.
- Server-side city production validation rejects buildings whose required city improvements are missing, independent of client-submitted payloads.
- Visibility legal actions and deterministic AI production choices filter buildings by required city improvements.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestRuntimeBuildingProductionRequiresPrerequisiteBuildings -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/ai ./internal/gameplay/visibility -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Full retail building placement constraints such as coastal-only, sea-city bans, and specialized city-type prerequisites remain incomplete.
