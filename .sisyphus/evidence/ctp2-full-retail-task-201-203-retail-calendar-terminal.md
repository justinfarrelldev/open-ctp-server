# CTP2 Full Retail Tasks 201-203: retail calendar and terminal score proof

## Source evidence
- `Const.txt` defines `END_OF_GAME_YEAR_EARLY_WARNING 2200` and `END_OF_GAME_YEAR 2300`.
- `DiffDB.txt` first difficulty `TIME_SCALE` starts at -4000 with periods 0/20, 150/10, 340/5, 500/2, and 600/1 years per turn.
- Original `NewTurnCount::RunNewYearMessages` compares `GetCurrentYear()` against `g_theConstDB->GetEndOfGameYear()` and ends the game by score when the end year is reached.

## Implementation
- Retail catalog loading now preserves calendar/end-year data and computes final score turn from the retail time scale (`2300 AD` => turn `1000`).
- Retail runtime projection now exposes the computed final turn and enables supported natural victory evaluators in deterministic order.
- Added a retail terminal proof that advances a retail-runtime state from turn 999 to 1000 and observes a `score_2300` terminal event/winner.

## Tests
- `go test ./internal/gameplay/rules -run 'TestLoadRetailRulesFromDirectoryCompilesTypedCatalogs|TestProjectRetailRuntimeIncludesCompleteReleasedCatalog' -count=1`
- `go test ./internal/e2e/fullgame -run TestRetailRuntimeScoreVictoryCanReachTerminalFinalTurn -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/e2e/fullgame -run 'TestRetail(RuntimeScoreVictoryCanReachTerminalFinalTurn|AIAutoPlayRecordsAtLeastHundredTurns)' -count=1`
