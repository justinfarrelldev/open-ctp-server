# CTP2 Full Retail Task 45: wonder building-everywhere effects

## Source references
- Retail `Wonder.txt` uses `BuildingEverywhere` and `ActualBuildingEverywhere` for wonders such as `WONDER_CENTRAL_MATTER_DECOMPILER`, `WONDER_HAGIA_SOPHIA`, `WONDER_INTERNET`, and `WONDER_DREAMLAND`.
- Retail `buildings.txt` defines the target building effects, e.g. `IMPROVE_COMPUTER_CENTER` has `SciencePercent 0.2` and `IMPROVE_MATTER_DECOMPILER` has `ProductionPollutionPercent -0.5`.

## Implementation
- Runtime wonder projection now inherits supported effects from target buildings named by `BuildingEverywhere` and `ActualBuildingEverywhere`.
- Inherited effects are emitted as normal safe runtime effects on the wonder, so existing city/wonder effect execution applies them globally through the established wonder effect path.
- Unsupported target building mechanics remain omitted until their building effect kinds are implemented.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- This projects supported effects from building-everywhere targets; it does not yet materialize actual city building inventory, unsupported happiness/defense/protection effects, or target building upkeep semantics.
