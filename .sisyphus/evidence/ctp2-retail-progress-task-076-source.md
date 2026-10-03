# Task 076 source check: active defense fires once per turn

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:940-941` skips possible active defenders already flagged with `k_UDF_USED_ACTIVE_DEFENSE`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:1013-1019` resolves active-defense attacks and then sets `k_UDF_USED_ACTIVE_DEFENSE` on the active defender.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:3264` begins unit turn refresh, and `:3291-3295` clears `k_UDF_USED_ACTIVE_DEFENSE` during `UnitData::BeginTurn`.

Implementation notes:

- Added canonical `Unit.UsedActiveDefense` state to prevent the same interceptor from firing against multiple bombarders in one turn.
- Active-defense use is set when an interceptor attempts active defense, regardless of hit result, matching the source flag placement.
- The flag resets in the server's turn refresh path together with movement refresh.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-076-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-076-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-076-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-076-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-076-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-076-diff-check.txt`
