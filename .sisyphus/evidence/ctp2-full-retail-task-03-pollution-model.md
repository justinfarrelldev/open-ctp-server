# CTP2 full-retail parity task 03 pollution model evidence

Scope: add a source-proven retail global-pollution model using parsed `pollution.txt` trigger lists.

Original source checked:
- `gs/gameobj/Pollution.cpp`:
  - `Pollution::AtTriggerLevel` compares `GetGlobalPollutionLevel()` with the current map-size trigger using strict `>` and suppresses disasters when the reduce-world-pollution wonder is active.
  - `Pollution::BeginTurn` applies the disaster for the current phase and then advances to the next trigger phase, capped at the last trigger.
  - `Pollution::EndRound` shifts the short global history and stores `history[0] + sum(player pollution)` as the new current value.
  - `Pollution::GetRoundsToNextDisaster` projects turns using `(trigger - current) / (current - previous)` and returns zero for flat/downward pollution.
- `gs/gameobj/PollutionConst.h`: global pollution history length is 5 turns.
- `gs/database/PollutionRecord.cpp` and `pollutiontoken.h`: trigger-list rows are repeated `POLLUTION_TRIGGER <level> <disaster>` entries.

Validation commands:
- `go test ./internal/gameplay/rules -run RetailPollution -count=1`
- `go test ./internal/gameplay/rules -count=1`
