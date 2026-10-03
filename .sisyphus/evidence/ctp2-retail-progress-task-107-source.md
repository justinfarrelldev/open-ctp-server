# Task 107 source evidence: effective-building economy effects

## Retail source checks

- `CityData::GetEffectiveBuildings()` returns actual city improvements OR the owner's wonder-granted building mask: `m_built_improvements | g_player[m_owner]->GetWonderBuildings()` (`ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:5717-5720`).
- Retail economy/yield helpers consume `GetEffectiveBuildings()` rather than only actual improvements:
  - Production percent: `buildingutil_GetProductionPercent(GetEffectiveBuildings(), prodBonus)` (`CityData.cpp:1193`).
  - Food percent: `buildingutil_GetFoodPercent(GetEffectiveBuildings(), foodBonus)` (`CityData.cpp:1470`).
  - Science percent: `buildingutil_GetSciencePercent(GetEffectiveBuildings(), s)` (`CityData.cpp:4566`).
  - Commerce/gold percent: `buildingutil_GetCommercePercent(GetEffectiveBuildings(), tradeBonus, m_owner)` (`CityData.cpp:5821`).
  - Gold per citizen: `buildingutil_GetGoldPerCitizen(GetEffectiveBuildings())` (`CityData.cpp:5833`).
- Released data grants effective buildings from wonders, for example `WONDER_GLOBAL_NANO_CLEANER` has `BuildingEverywhere IMPROVE_MATTER_DECOMPILER` (`ctp2_released_game/ctp2_data/default/gamedata/Wonder.txt:25`), and several other wonders use `ActualBuildingEverywhere`.

## Server implementation

- `cityLocalEffects` now augments actual city building effects with active owner/global `building_everywhere` wonder building effects looked up from pinned runtime production definitions.
- Effective building effects are applied through the same city-local effect path used by city yields, protection/combat helpers, and research effects.
- Runtime-imported wonders already embed some building effects for compatibility; the effective-building lookup skips identical embedded wonder effects so imported retail wonders do not double count.
- Effective buildings do not add building upkeep, matching the retail effective-building bitmask behavior rather than constructing an actual city improvement.

## Validation

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-107-red.txt`
- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-107-green-target.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-107-green.txt`
- Freeze RED/GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-107-freeze-red.txt`, `.sisyphus/evidence/ctp2-retail-progress-task-107-freeze-green.txt`
- Fullgame GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-107-fullgame-green.txt`
