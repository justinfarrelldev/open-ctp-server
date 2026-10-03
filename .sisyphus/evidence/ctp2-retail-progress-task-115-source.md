# Task 115 source evidence: production/prohibition wonder effects and reversal

## Retail source checks

- Retail parses `ReformCities`, `FreeSlaves`, `PollutersToParks`, `EnableParkRangers`, `EliminateNukes`, and `ProhibitSlavers`-style prohibition fields in `WonderRecord` (`ctp2_source_code/ctp2_code/gs/newdb/WonderRecord.cpp:216-235`, `575-639`).
- Retail utility helpers read these effects from live built-wonder bitsets (`wonderutil.cpp:85-122`, `175-177`).
- On wonder completion retail runs one-shot effects:
  - `PollutersToParks` converts polluted cities/tiles (`Player.cpp:7441-7450`).
  - `FreeSlaves` frees slaves in all cities (`Player.cpp:7495-7499`).
  - `EliminateNukes` iterates cities and calls `CityData::EliminateNukes()` (`Player.cpp:7552`, `CityData.cpp:4870-4880`), which removes nuclear plants and illegal queue items.
- Retail production/legality paths consult live global built-wonder state, so destroyed/obsolete wonders stop applying reversible production prohibitions:
  - Free-slave checks use `wonderutil_GetFreeSlaves(g_theWonderTracker->GetBuiltWonders())` (`CityData.cpp:4185`, `Player.cpp:10209`).
  - Other wonder helpers likewise read current built-wonder bitsets rather than permanent per-nation flags.

## Server implementation

- Existing coverage executes `EnableParkRangers`, `FreeSlaves`, `ProhibitSlavers`, `EliminateNukes`, `ReformCities`, and `PollutersToParks`, including illegal queue cleanup and production blocking.
- Added `NuclearAttacksEliminated` so active `EliminateNukes` wonders remain reversible: active wonders block nuke production/actions; destroyed/obsolete tracked wonders no longer do, while legacy nation markers are honored only when no wonder record exists.
- Updated production and visibility/legal-action suppression to use the reversible helper.
- Added `TestRetailDestroyedEliminateNukesWonderReversesProductionBlock` to lock the destroyed-wonder reversal edge case.

## Validation

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-115-red.txt`
- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-115-green-target.txt`
- Visibility GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-115-visibility-green.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-115-green.txt`
- Freeze RED/GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-115-freeze-red.txt`, `.sisyphus/evidence/ctp2-retail-progress-task-115-freeze-green.txt`
- Fullgame GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-115-fullgame-green.txt`
