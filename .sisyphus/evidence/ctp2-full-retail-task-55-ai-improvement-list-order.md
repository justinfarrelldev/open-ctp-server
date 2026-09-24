# CTP2 Full Retail Task 55: AI terrain-improvement list order

## Source references
- Retail `ctp2_data/default/aidata/ImprovementLists.txt` orders terrain improvements, including farms before advanced farms in `IMPROVEMENT_LIST_GROWTH`.
- The runtime already imports and projects `ImprovementLists.txt` as safe public AI strategy metadata.

## Implementation
- Viewer legal action generation now orders projected terrain-improvement actions by retail AI improvement-list order before falling back to canonical runtime order.
- The deterministic AI planner now consumes projected build-improvement legal actions and scores them by retail improvement-list order.
- Pollution-reducing improvement actions are filtered to polluted visible tiles so synthetic cleanup scenarios keep the prior bounded cleanup behavior.

## Tests
- `go test ./internal/gameplay/ai -count=1`
- `go test ./internal/gameplay/visibility -run TestTerrainActionsUseRetailAIImprovementListOrder -count=1`
- `go test ./internal/e2e/fullgame -run TestTaskFiveExecutionSemanticsRemainGreen -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- AI does not yet perform full retail worker assignment, terrain valuation, or multi-turn public-works strategy; action legality remains bounded by visible projection data and canonical executor validation.
