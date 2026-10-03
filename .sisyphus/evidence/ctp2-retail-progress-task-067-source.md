# Task 067 source check: retail flanker row targeting

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CTP2Combat.cpp:149-156` classifies unit combat type as `UNIT_TYPE_FLANKER` when the unit record has `IsFlanker`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CTP2Combat.cpp:303-336` implements `CombatField::FindTargetForAttackFrom`; melee attacks first target the same row, and only flanker attackers scan nearby rows above/below when the same row is empty.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CTP2Combat.cpp:1197-1214` passes `att->GetCombatType() == UNIT_TYPE_FLANKER` into `FindTargetForAttackFrom` during melee attacks.

Implementation notes:

- Melee strike target selection now preserves row semantics: non-flankers do not retarget from an empty same row, while flankers target the nearest occupied compatible row.
- Existing domain-mask behavior is preserved when a same-row defender exists but is incompatible: the attacker can still select a compatible living target rather than stalling on an impossible target.
- Ranged attacks remain on the existing ranged target path; this slice covers `IsFlanker` melee targeting only.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-067-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-067-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-067-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-067-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-067-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-067-diff-check.txt`
