# CTP2 Full Retail Task 49: soothsaying unit projection

## Source references
- Retail `Units.txt` defines `CanSoothsay` blocks on units such as `UNIT_CLERIC`.
- The server already has bounded soothsaying command execution via `special_soothsaying`; this task wires the retail unit field into runtime effects.

## Implementation
- Runtime unit projection now maps retail `CanSoothsay` to safe `special_soothsaying` unit effects.
- Viewer projections and legal action generation already consume `special_soothsaying`, so retail cleric-style units can expose and execute the bounded soothsaying action without hidden-state leakage.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Full retail soothsaying probability/effect payloads and related religious/happiness consequences remain incomplete.
