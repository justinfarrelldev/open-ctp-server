# Task 070 source check: LossMoveToDmgTwo movement refresh

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:594-604` implements `UnitData::ResetMovement`: `LossMoveToDmgNone` refreshes full max movement; `LossMoveToDmgTwo` refreshes to `max(2.0f, maxMove * hp / maxHp)`; default damaged units refresh proportionally to HP.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/UnitRecord.cpp:548-549` and `:1352-1356` parse both `LossMoveToDmgNone` and `LossMoveToDmgTwo` unit flags.

Implementation notes:

- Retail runtime projection now emits `loss_move_to_dmg_two` for units flagged `LossMoveToDmgTwo`.
- Movement refresh now applies the retail minimum-2 movement floor for damaged units with `loss_move_to_dmg_two`; `loss_move_to_dmg_none` still takes precedence and grants full movement.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-070-red.txt`
- GREEN targeted: `.sisyphus/evidence/ctp2-retail-progress-task-070-green-targeted.txt`
- GREEN package: `.sisyphus/evidence/ctp2-retail-progress-task-070-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-070-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-070-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-070-diff-check.txt`
