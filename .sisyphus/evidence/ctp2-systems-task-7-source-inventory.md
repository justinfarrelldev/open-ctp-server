# Task 7 CTP2 source inventory: city economy, production, buildings, wonders

Source root inspected before and after implementation: `/home/justi/Projects/open-ctp-server/ctp2/ctp2_source_code`.

## Production queues and unlocks
- `ctp2_code/gs/gameobj/CityData.cpp` `CityData::BuildUnit`, `BuildImprovement`, and `BuildWonder` enqueue unit/building/wonder work with shield/production costs after `CanBuild*` and advance checks.
- Server alignment: city production commands validate ownership, prerequisites, duplicate queue entries, and wonder uniqueness; round recalculation spends stored production and completes queued buildings/wonders deterministically.

## Building and wonder effects
- `ctp2_code/gs/gameobj/CityData.cpp` uses built improvements and wonders in helpers such as `buildingutil_GetFoodPercent`, `buildingutil_GetProductionPercent`, and `wonderutil_GetDecreaseMaintenance`.
- Server alignment: synthetic `Effect` records are bounded and applied through `cityEffects` from city, nation, building, and wonder scopes without copying retail modifiers.

## Upkeep/failure semantics
- `ctp2_code/gs/gameobj/CityData.cpp` `SupportBuildings` computes building upkeep, subtracts it from city trade, and sells buildings when maintenance cannot be paid.
- Server alignment: the current deterministic slice applies upkeep as a production-yield reduction and clamps negative production to zero; forced sale/rush-buy retail behavior is intentionally deferred.

## Terrain base yields
- `ctp2_code/gs/newdb/terrain.cdb` defines terrain `Gold`, `Food`, and `Production` values.
- Server alignment: synthetic terrain types have deterministic food/production/gold base yields and are recalculated into canonical city yield state.
