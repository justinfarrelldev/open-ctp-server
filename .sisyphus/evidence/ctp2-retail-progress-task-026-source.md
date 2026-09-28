# Task 026 source evidence: conquest capture routing

Retail CTP2 mechanics checked for the capture-route recovery fix:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:5195` calls `IsMovePointsEnough(newPos)` before moving, so the runner must only submit current public/legal move steps and must wait for turns when the public projection has no move action.
- `ctp2/ctp2_source_code/ctp2_code/gs/world/cellunitlist.cpp:547-577` checks every unit's remaining movement points against terrain cost, matching the projected first-step legality filter in the test harness.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:5342-5404` executes a successful move through `MoveIntoCell`/`FinishMove`, then removes units that cannot legally occupy the destination. The harness therefore keeps capture movement on public legal actions rather than forcing planner-only destinations.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:5394-5411` (`FinishAttack`) sets non-air attackers' movement points to zero after combat before the victory move, so the harness must preserve non-combat reserve units and be able to switch to a visible mobile capture-capable reserve after combat.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:7502-7531` (`CheckWasEnemyVisible`) gates enemy contact/attack handling on actual visibility, supporting the projection-only rule: remembered/fog state may guide planning, but emitted actions must be current public legal actions.
- `ctp2/ctp2_source_code/ctp2_code/gs/world/WrlUnit.cpp:512-518` considers cities with `GetEverVisible()` for radar/top-unit display, matching the use of remembered city/tile data only as planning hints.

Implementation consequence: known/stale targets can guide planning through `PlanKnown`, but every submitted capture step is checked against public `LegalActions`; generic exploration uses visible reachable frontiers/backtracking to avoid stale-memory oscillation, while known city targets can use remembered routes for one legal first step at a time.
