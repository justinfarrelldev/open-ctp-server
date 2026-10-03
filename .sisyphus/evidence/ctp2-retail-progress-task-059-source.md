# Task 059 source check: wonder BuildingEverywhere effective buildings

Retail data/source evidence:

- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Wonder.txt:25` gives `WONDER_CENTRAL_MATTER_DECOMPILER` `BuildingEverywhere IMPROVE_MATTER_DECOMPILER`.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Wonder.txt:194`, `217`, and `251` use `ActualBuildingEverywhere` for Shrine, Computer Center, and Force Field style wonders.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:7655-7658` sets the owner's `m_wonderBuildings` bit when a wonder has `BuildingEverywhere`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:7688-7690` clears that bit when the wonder is removed.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:5719-5722` returns `m_built_improvements | g_player[m_owner]->GetWonderBuildings()` as city effective buildings.

Implementation notes:

- Retail runtime now projects a `building_everywhere` marker for both released `BuildingEverywhere` and `ActualBuildingEverywhere` wonder fields while preserving the already-projected building effects.
- Production prerequisite checks now use effective buildings: real city buildings plus active owner/global wonder `building_everywhere` markers.
- Destroyed, obsolete, or foreign owner-only wonders do not satisfy another city's prerequisite building.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-059-red.txt`
- GREEN engine: `.sisyphus/evidence/ctp2-retail-progress-task-059-green-engine.txt`
- GREEN rules: `.sisyphus/evidence/ctp2-retail-progress-task-059-green-rules.txt`
- GREEN packages: `.sisyphus/evidence/ctp2-retail-progress-task-059-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-059-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-059-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-059-diff-check.txt`
