# CTP2 full-retail parity task 18 AI autoplay 100-turn harness evidence

Scope: add the required long-run retail runtime autoplay regression harness and generate a visual transcript artifact.

Implemented server-compatible slice:
- Added `TestRetailAIAutoPlayRecordsAtLeastHundredTurns`.
- The harness loads user-owned released-game gamedata, projects/pins retail runtime rules, creates four AI players, gives each a retail starter unit, and drives turns using `ai.Plan` from `visibility.Project` only.
- The test reaches at least 100 completed turn records and writes JSONL turn summaries containing turn, active actor, submitted command types, city/unit/advance/wonder counts, pollution count, diplomacy records, and canonical state hash.
- By default the transcript goes to `t.TempDir`; setting `OPEN_CTP_RECORD_RETAIL_AI_TURNS=<path>` records a persistent artifact.

Generated artifact:
- `.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl`
- 100 JSONL records.
- First turn record reaches turn 2 after `ai-1` founds a city, selects research/government, queues production, and closes the turn.
- Final record reaches turn 101.

Validation commands:
- `go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

Limitations noted for future parity work:
- This is a 100-turn regression harness, not full game completion yet.
- Current partial retail mechanics produce limited midgame progress; later tasks must make this transcript show full research/economy/military/diplomacy/pollution progression and eventually terminal victory/final-turn completion.
