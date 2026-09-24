# CTP2 Full Retail Task 48: city defense building effects

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/building.cdb:14` defines `DefendersPercent`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/buildingutil.cpp:110-121` accumulates `DefendersPercent` for built city improvements.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:3593-3610` applies city defender bonuses when a unit is defending in a city.
- Retail `buildings.txt` gives `IMPROVE_CITY_WALLS` `DefendersPercent 15`.

## Implementation
- Runtime building projection maps `DefendersPercent` to safe `combat_defense` effects.
- Combat defense modifiers for a unit in a city now use full city effects, including non-obsolete building effects and applicable wonder/nation effects, rather than only direct city effects.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestRetailCityDefenseBuildingEffectsModifyDefenders -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Domain-specific city offense bonuses (`OffenseBonusLand/Water/Air`), wall-ignore behavior, and full combat formula parity remain incomplete.
