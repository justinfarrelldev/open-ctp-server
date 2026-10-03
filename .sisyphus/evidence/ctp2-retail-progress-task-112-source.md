# Task 112 source evidence: effective-building brokerage closure

## Retail source checks

- Retail commerce-percent calculation is effective-building-aware: `CityData.cpp:5821` calls `buildingutil_GetCommercePercent(GetEffectiveBuildings(), tradeBonus, m_owner)`.
- `buildingutil_GetCommercePercent` reads each building's commerce percent and, when `GetBrokerage()` is set, increases that building's commerce percent by `wonderutil_GetIncreaseBrokerages(...) / 100.0` before adding it to the city total (`buildingutil.cpp:387-405`).
- `IncreaseBrokerages` is a wonder field parsed in `WonderRecord` (`ctp2_source_code/ctp2_code/gs/newdb/WonderRecord.cpp:268-269`, `750-752`).
- Because retail passes `GetEffectiveBuildings()` into the commerce helper, brokerage buildings granted by `BuildingEverywhere`/`ActualBuildingEverywhere` must also receive the `IncreaseBrokerages` amplification.

## Server implementation

- Added grouped effective-building effects so building-local relationships are preserved instead of relying only on flattened city effects.
- `brokerageGoldPercentBonus` now evaluates actual building effect groups plus wonder-granted effective-building effect groups, so only commerce percent on brokerage buildings receives the `IncreaseBrokerages` multiplier.
- Existing flattened effective-building behavior for yields/science/combat/protection remains unchanged; `effectiveBuildingEffects` now flattens the same grouped helper.

## Validation

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-112-red.txt`
- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-112-green-target.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-112-green.txt`
- Freeze RED/GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-112-freeze-red.txt`, `.sisyphus/evidence/ctp2-retail-progress-task-112-freeze-green.txt`
- Fullgame GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-112-fullgame-green.txt`
