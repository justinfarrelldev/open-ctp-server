# CTP2 Full Retail Task 50: city offense building effects

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/building.cdb:51-53` defines `OffenseBonusLand`, `OffenseBonusWater`, and `OffenseBonusAir`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:5701-5714` applies offense bonuses from city buildings according to defender movement/domain.
- Retail `buildings.txt` uses these fields on `IMPROVE_BALLISTA_TOWERS`, `IMPROVE_BATTLEMENTS`, and `IMPROVE_FLAK_TOWERS`.

## Implementation
- Runtime building projection maps city offense bonuses to domain-specific safe effects: `combat_attack_land`, `combat_attack_sea`, and `combat_attack_air`.
- Combat now applies city-local attack effects when the attacker is on a city tile, matching the CTP2 city offense bonus shape.
- City combat effects now use local city/building/wonder effects, avoiding double-application of nation effects that are already applied at the unit/nation layer.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetailCityOffenseBuildingEffectsModifyAttackersByDefenderDomain|TestRetailCityDefenseBuildingEffectsModifyDefenders|TestCombatVectorsResolveAtomicallyWithSourceOrderAndCapture' -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Full combat formula parity, wall-ignore behavior, and many unit-specific combat bonuses remain incomplete.
