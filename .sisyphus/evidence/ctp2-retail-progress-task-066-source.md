# Task 066 source check: city bombard wall defense uses actual improvements

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:4278-4283` computes undefended city bombard probability as `rec->GetZBRangeAttack() - buildingutil_GetCityWallsDefense(c.GetCityData()->GetImprovements())`.
- That source uses the city's actual built improvements (`GetImprovements()`), not effective/wonder/global local effects.

Implementation notes:

- Undefended-city bombard now subtracts wall defense only from actual city buildings carrying `city_walls` effects.
- Global/wonder/direct city wall effects still apply to normal combat city defense through the existing effective city-defense path, but no longer reduce city bombard building/population probabilities.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-066-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-066-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-066-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-066-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-066-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-066-diff-check.txt`
