# Task 111 source evidence: building military effects

## Retail source checks

- Retail parses `DefendersPercent`, `CityWalls`, `OffenseBonusLand`, `OffenseBonusWater`, `OffenseBonusAir`, and `Airport` (`ctp2_source_code/ctp2_code/gs/newdb/BuildingRecord.cpp:214-215`, `270-275`, `393`, `430-438`).
- Defender/city-wall helpers use effective buildings:
  - `buildingutil_GetDefendersBonus(GetEffectiveBuildings(), ...)` is used from `CityData` defensive-bonus refresh paths (`CityData.cpp:2237`, `2674`, `2696`, `4783`, `5089`, `5633`).
  - `CityData::HasCityWalls()` returns `buildingutil_GetCityWalls(GetEffectiveBuildings())` (`CityData.cpp:3038-3040`).
  - `UnitData` applies city defensive bonus and uses city-wall defense/reduction logic around attacker wall-reduction effects (`UnitData.cpp:3536-3547`).
- Offensive city-building bonuses use effective buildings by defender domain: `CityData::GetOffenseBonus` adds land/water/air building utility values from `GetEffectiveBuildings()` (`CityData.cpp:5701-5714`).
- Retail `Airport` is parsed and exposed via `CityData::HasAirport()` using `GetEffectiveBuildings()` (`CityData.cpp:3033-3035`), but source call sites are output/visibility/UI and AI retreat helpers; fuel refuel itself happens in any city or owned airfield installation in `UnitData::CheckForRefuel()` (`UnitData.cpp:5662-5689`).
- The released source has an obvious copy/paste bug in `buildingutil_GetOffenseBonusWater` and `buildingutil_GetOffenseBonusAir` where both read `GetOffenseBonusLand`; the server intentionally maps released data's water/air fields to water/air domain attack effects.

## Server coverage

- `DefendersPercent` imports to city/building combat-defense effects and `CityWalls` to the wall marker used by ignore/reduce-city-wall logic.
- `OffenseBonusLand`, `OffenseBonusWater`, and `OffenseBonusAir` import to domain-specific city attack effects.
- Direct building paths were already covered by combat tests; `TestRetailEffectiveBuildingMilitaryEffectsModifyCombat` now locks wonder-granted effective-building attack and city-wall defense/ignore behavior.
- `Airport` remains an imported marker; server city refuel already matches the retail `CheckForRefuel` behavior for owned cities and tile airbase overlays.

## Validation

- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-111-target-green.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-111-green.txt`
