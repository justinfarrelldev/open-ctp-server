# Task 113 source evidence: numeric wonder effects

## Retail source checks

- Retail parses the numeric wonder fields in `WonderRecord`: `DecCrimePercent`, `IncKnowledgePercent`, `DecEmpireSize`, `IncHappinessEmpire`, route/television gold fields, `IncreaseProduction`, `BonusGold`, and `IncreaseFoodAllCities` (`ctp2_source_code/ctp2_code/gs/newdb/WonderRecord.cpp:207-258`).
- Retail wonder utility accessors aggregate these fields from active built-wonder bitsets (`wonderutil.cpp:132-172`, `265-272`).
- Retail production and food city calculations apply `IncreaseProduction` and `IncreaseFoodAllCities` from the owner's built wonders (`CityData.cpp:1192-1201`, `1468-1477`).
- Retail treasury calculations apply water-route, international-route, television, and flat bonus-gold wonder fields (`Player.cpp:2330-2368`).
- Server-side gating mirrors the retail active-wonder model by excluding destroyed and obsolete wonders from `ownerWonderEffectSum`/`wonderAppliesToPlayer`.

## Server coverage

- Runtime import maps numeric wonder fields to city/research/happiness/crime/empire-size/gold effects.
- City yield/research/happiness/crime/distance/gold paths consume those effects through active owner/global wonder helpers, with destroyed and obsolete wonders filtered out.
- Added `TestRetailNumericWonderEffectsRespectDestructionAndObsolescence` to lock active-wonder gating for numeric effects; existing wonder-gold/yield/brokerage tests cover the per-field execution paths.

## Validation

- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-113-target-green.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-113-green.txt`
