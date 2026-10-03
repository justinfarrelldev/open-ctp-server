# Task 116 source evidence: happiness/readiness/pollution wonder effects

## Retail source checks

- Retail parses `ReduceReadinessCost`, `TemporaryFullHappiness`, and `NoPollutionUnhappiness` in `WonderRecord` (`ctp2_source_code/ctp2_code/gs/newdb/WonderRecord.cpp:204-205`, `259-262`, `534-536`, `720-731`).
- Retail `TemporaryFullHappiness` is applied on wonder completion to every city owned by the wonder builder via `SetFullHappinessTurns(fullHappinessTurns)` (`Player.cpp:7524-7531`).
- Retail `NoPollutionUnhappiness` suppresses pollution happiness penalties in `Happy::CalcPollution` (`Happy.cpp:359-363`).
- Retail `ReduceReadinessCost` triggers readiness-cost recalculation for the owner when the wonder state changes (`Player.cpp:9891-9892`) and is exposed through `wonderutil_GetReduceReadinessCost` (`wonderutil.cpp:127`).

## Server coverage

- `TemporaryFullHappiness` writes expiring city markers for owner cities and lifecycle ticking removes them after the configured duration.
- `NoPollutionUnhappiness` suppresses city pollution-unhappiness penalties through the active wonder helper.
- `ReduceReadinessCost` reduces support/readiness cost, stacks active owner wonders, ignores destroyed/foreign wonders, and caps at 100%.
- Added `TestRetailReduceReadinessCostWonderStacksAndCaps` to lock stacking/capping semantics.

## Validation

- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-116-target-green.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-116-green.txt`
