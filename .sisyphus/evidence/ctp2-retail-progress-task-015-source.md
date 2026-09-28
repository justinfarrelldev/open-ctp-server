# Task 015 source check: AI bombard legal actions

Retail CTP2 AI uses bombard-capable units explicitly:

- `ctp2/ctp2_source_code/ctp2_code/ai/ctpai.cpp::CtpAi::BombardAdjacentEnemies` checks `army->CanBombard()` and emits `GEV_BombardOrder` when adjacent bombard targets are valid.
- `ctp2/ctp2_source_code/ctp2_code/ai/ctpai.cpp` calls `CtpAi::BombardAdjacentEnemies(army)` during AI army processing.
- `ctp2/ctp2_source_code/ctp2_code/robotcom/goals/GoalBombard.cpp` and `robotcom/Agent/armyagent.cpp` model bombard goals and `CanBombardTarget` checks for non-adjacent planner bombardment.

Implementation notes:

- Engine and visibility already expose legal `bombard` actions with unit and target-tile payloads.
- The AI planner now consumes those projected legal bombard actions instead of ignoring bombard-capable units, preserving the public-projection-only planning boundary.
