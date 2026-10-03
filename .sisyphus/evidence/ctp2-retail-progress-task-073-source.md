# Task 073 source check: active-defense range is measured to attacker

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:943-950` reads a possible active defender's `GetActiveDefenseRange()` and compares distance from the active defender to `m_array[0]`, the attacking army/unit, via `UnitData::GetDistance(input[i].AccessData(), m_array[0].AccessData(), ...)`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:977-989` searches possible defenders around the target position first, then `GetActiveDefenders` applies the actual defender-to-attacker range check.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:1013-1018` runs active-defense bombardment for eligible active defenders.

Implementation notes:

- Active-defense interception now checks interceptor range to the bombarding unit's origin tile rather than to the target tile.
- Existing domain and enabler cargo checks remain unchanged.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-073-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-073-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-073-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-073-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-073-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-073-diff-check.txt`
