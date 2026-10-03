# Task 063 source check: capture resets city special/timed effects

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:3587-3638` implements `CityData::ResetCityOwner` during owner changes.
- The same source block clears city-owner-specific state: `m_franchiseTurnsRemaining = 0`, `m_franchise_owner = -1`, `m_watchfulTurns = 0`, bio/nano infection owner/timers, conversion-to-self, happiness attacked, and `m_isInjoined = FALSE`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:3639` calls `m_happy->ClearTimedChanges()`, removing timed happiness adjustments on capture.

Implementation notes:

- `captureCityAt` now clears owner-specific city effects (`franchise`, `injunction`, `watchful`) and timed effects (`Turns > 0`) before owner transfer.
- Non-timed base effects remain intact.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-063-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-063-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-063-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-063-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-063-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-063-diff-check.txt`
