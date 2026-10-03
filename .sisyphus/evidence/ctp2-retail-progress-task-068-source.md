# Task 068 source check: retail ranged row targeting

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CTP2Combat.cpp:254-299` implements `CombatField::FindTargetForRangedAttackFrom`; ranged attacks scan from the attacker's row, then nearby rows by increasing distance, and only fall back to the first active target if no row-near target was found.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CTP2Combat.cpp:1094-1114` calls `FindTargetForRangedAttackFrom(x, y, &dx, &dy)` for each ranged attacker.

Implementation notes:

- Ranged combat target selection now prefers the same tactical row, then nearest rows above/below, before falling back to the first compatible living target.
- This builds on Task 067's melee row-targeting work while preserving compatibility/domain filtering.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-068-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-068-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-068-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-068-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-068-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-068-diff-check.txt`
