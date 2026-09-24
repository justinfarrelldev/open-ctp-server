# CTP2 Full Retail Task 54: mounted defense bonuses

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/unit.cdb:117,131,380` defines `IsMounted`, `BonusAgainstMounted`, and `MountedBonus`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:3612-3614` adds defender mounted-bonus defense when the attacker is mounted.
- Retail `Units.txt` marks `UNIT_CAVALRY` as `IsMounted` and `UNIT_PIKEMEN` as `BonusAgainstMounted` with `MountedBonus 0.5`.

## Implementation
- Runtime unit projection now maps mounted attackers to a safe `mounted` effect.
- Runtime unit projection maps anti-mounted unit records to `combat_defense_vs_mounted_percent` effects.
- Combat defense resolution applies the defender's percent bonus against attackers carrying the `mounted` effect.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetailMountedDefenseBonusModifiesDefenderAgainstMountedAttackers|TestCombatVectorsResolveAtomicallyWithSourceOrderAndCapture' -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Air-defense bonus, flanking, veteran modifiers, and the full CTP2 combat stack remain incomplete.
