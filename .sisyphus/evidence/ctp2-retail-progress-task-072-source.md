# Task 072 source check: counter-bombard honors defender bombard range

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:1263-1273` implements `UnitData::CanBombard`; it checks target distance against `rec->GetBombardRange() + 0.5` before evaluating target movement domains.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:1298-1304` implements `CanCounterBombard` by requiring `GetCanCounterBombard()` and then delegating to `CanBombard(defender)`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:4393-4395` calls `defender[i].CanCounterBombard(*this)` before counter-bombardment.

Implementation notes:

- Counter-bombard now requires the defender's own `BombardRange` to reach the original bombarder.
- Domain filtering from Task 071 remains intact through the defender's own bombard target rules.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-072-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-072-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-072-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-072-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-072-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-072-diff-check.txt`
