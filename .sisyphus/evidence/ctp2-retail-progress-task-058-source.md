# Task 058 source check: veteran combat coefficient

Retail data/source evidence:

- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Const.txt:133` defines `VETERAN_COEF 50` as the percent improvement for veteran units.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:1397-1398` increases bombard hit probability by `g_theConstDB->GetVetCoef()` when the bombarding unit is veteran.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:1421-1422` applies the same veteran coefficient to one-round bombard hit probability.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:1503-1505` increases melee attack strength by the veteran coefficient before combat odds are computed.

Implementation notes:

- Runtime constants now import/project `VETERAN_COEF` as `veteran_coef_percent`.
- Combat attack odds now apply the veteran coefficient after base attack plus attack modifiers, matching retail's `GetAttack(...)` then veteran scaling order; ranged bombard/active-defense/counter-bombard odds use veteran-adjusted ranged strength.
- The fallback coefficient remains the released retail value, 50%, for nil/minimal runtimes.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-058-red.txt`
- GREEN engine: `.sisyphus/evidence/ctp2-retail-progress-task-058-green-engine.txt`
- GREEN packages: `.sisyphus/evidence/ctp2-retail-progress-task-058-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-058-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-058-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-058-diff-check.txt`
