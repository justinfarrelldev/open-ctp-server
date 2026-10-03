# Task 077 source check: active-defense interceptor ordering

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:952-963` inserts active defenders into the output list before lower-attack defenders using `input[i].GetAttack() > output[j].GetAttack()`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:1013-1019` iterates that ordered active-defender list and stops later active defense only when the attacking army/unit has been destroyed.

Implementation notes:

- Active-defense interceptors now sort by descending `Attack`, with ID as a deterministic tie-breaker.
- This preserves retail's stronger-interceptor-first behavior and prevents weaker lower-priority interceptors from consuming their once-per-turn use when a stronger interceptor destroys the attacker first.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-077-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-077-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-077-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-077-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-077-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-077-diff-check.txt`
