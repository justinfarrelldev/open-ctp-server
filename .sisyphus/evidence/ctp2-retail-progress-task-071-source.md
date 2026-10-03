# Task 071 source check: counter-bombard considers every defender in target stack

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:4365-4371` runs bombard attacks for the attacking army against the target defender stack.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:4393-4395` then iterates `for (i = 0; i < defender.Num(); i++)` and lets every defender unit that `CanCounterBombard(*this)` bombard the attacking army.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:1298-1304` implements `CanCounterBombard` by checking the defender's `CanCounterBombard` flag and its own `CanBombard(defender)` result.

Implementation notes:

- Counter-bombard now considers all living units on the target tile, not just units that were eligible targets for the attacker's bombard filters.
- This lets a counter-bombarder that the attacker could not bombard still fire if its own counter-bombard/bombard target rules allow attacking the bombarder.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-071-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-071-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-071-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-071-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-071-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-071-diff-check.txt`
