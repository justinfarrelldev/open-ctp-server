# Task 064 source check: undefended city bombard building/population damage

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:4314-4329` lets bombardment target an undefended city when the target tile has no defending army: `if (defender.Num() < 1) { if(BombardCity(point, TRUE)) return ORDER_RESULT_SUCCEEDED; }`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:4215-4299` implements `ArmyData::BombardCity`; it rejects owned/cannot-be-attacked cities, checks compatible bombard target movement, consumes the attack, then applies building destruction and population-loss rolls.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:4278-4286` computes city bombard probability as ranged bombard attack minus city-wall defense, clamps below zero, and rolls against `BombardDestroyBuildingChance()` and `BombardKillPopChance()`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:4746-4779` implements `DestroyRandomBuilding`, excluding capitol-designating buildings from bombard destruction.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Const.txt:283` and `:304` set `BOMBARD_DESTROY_BUILDING_CHANCE 0.1` and `BOMBARD_KILL_POP_CHANCE 0.1`.

Implementation notes:

- Retail constants now import/project `BombardDestroyBuildingPercent` and `BombardKillPopulationPercent` as percent points (retail 10/10).
- Unit bombardment can now target a visible enemy city with no unit targets; building destruction excludes `capitol` buildings and population loss cannot reduce population below 1.
- Public legal-action projection now includes non-stale visible enemy city bombard targets and suppresses stale city-memory targets.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-064-red.txt`
- GREEN targeted: `.sisyphus/evidence/ctp2-retail-progress-task-064-green-targeted.txt`
- GREEN package: `.sisyphus/evidence/ctp2-retail-progress-task-064-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-064-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-064-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-064-diff-check.txt`
