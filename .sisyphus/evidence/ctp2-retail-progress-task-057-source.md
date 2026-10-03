# Task 057: retail Gaia countdown prerequisites

## Retail source checked

- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/EndGameObjects.txt:8-30`
  - Gaia requires `MinNeeded 10` Gaia computers, `MinNeeded 10` power satellites, `MinNeeded 5` processing towers, and `MinCoverage 0.6`; the Solaris project activates after 10 turns.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/gaiacontroller.cpp:813-828`
  - `CanStartCountdown()` requires tower coverage, minimum towers, minimum mainframes, and minimum satellites.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/gaiacontroller.cpp:875-895`
  - `StartCountdown()` starts or cancels the completion turn depending on `CanStartCountdown()`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/gaiacontroller.cpp:174-210`, `739-779`
  - Tower coverage is recomputed from completed processing towers and the tower radius scales from min to max radius using satellite count.

## Implementation notes

- `RuntimeGaia` now carries retail object counts, tower coverage percentage, and tower radius parameters.
- Retail runtime projection populates the Gaia computer, power satellite, processing tower, coverage, and radius requirements from released retail values.
- Countdown/victory prerequisite checks now require the configured building counts, tower count, tower coverage, and Solaris wonder.
- Synthetic/minimal Gaia runtimes still default to one building plus one wonder for existing non-retail tests.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-057-red.txt`
- Green targeted tests:
  - `go test ./internal/gameplay/engine -run 'TestRetailGaiaCountdownRequiresConfigured|TestRuntimeRulesGaia' -count=1`
  - `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeIncludesCompleteReleasedCatalog -count=1`
