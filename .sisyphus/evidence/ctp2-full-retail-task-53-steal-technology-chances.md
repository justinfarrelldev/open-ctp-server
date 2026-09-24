# CTP2 Full Retail Task 53: steal-technology special-action chances

## Source references
- Retail `Units.txt` defines `StealTechnology` nested fields such as `RandomChance 0.5` and `DeathChance 0.25` on units including `UNIT_CYBER_NINJA`.
- The bounded engine already executes `special_steal` with percent chances and `special_capture` failure-removal chances; this task wires the retail nested values into those effects.

## Implementation
- Runtime unit projection now reads `StealTechnology.RandomChance` into `special_steal` percent values.
- Runtime unit projection now reads `StealTechnology.DeathChance` into `special_capture` percent values.
- Missing chance data still falls back to the previous 100% supported-action behavior for safety on synthetic data.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Specific targeted steal mode, veteran/elite chance variants, diplomatic consequences, and many additional special-agent actions remain incomplete.
