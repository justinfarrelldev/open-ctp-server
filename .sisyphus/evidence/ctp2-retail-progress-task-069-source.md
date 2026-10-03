# Task 069 source check: bombard hostilities start war and cancel trade

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:4248-4250` verifies city bombard attacks with `VerifyAttack(UNIT_ORDER_BOMBARD, point, c.GetOwner())`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:4344-4347` verifies unit bombard attacks with `VerifyAttack(UNIT_ORDER_BOMBARD, point, defender.GetOwner())`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:5676-5702` allows attacks only against enemies without blocked treaty violations, so a valid bombard is a hostile/war action in retail flow.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:4289-4296` and `:4352-4358` log cease-fire violation events for bombardment against the defender.

Implementation notes:

- Bombarding an enemy unit or undefended enemy city now sets war state between the attacker and defender owners.
- The existing war-state helper removes trade routes between combatants, matching other hostile combat transitions in this server.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-069-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-069-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-069-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-069-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-069-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-069-diff-check.txt`
