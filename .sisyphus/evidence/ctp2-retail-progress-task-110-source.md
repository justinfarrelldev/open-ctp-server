# Task 110 source evidence: building protection effects

## Retail source checks

- Retail parses protection building fields in `BuildingRecord`: `PreventConversion`, `PreventSlavery`, `ProtectFromNukes`, `ProtectFromBioAgents`, `ProtectFromNanoVirus`, and `ForceField` (`ctp2_source_code/ctp2_code/gs/newdb/BuildingRecord.cpp:238-255`).
- Retail protection helpers use `GetEffectiveBuildings()`:
  - `CityData::GetProtectFromNukes()` calls `buildingutil_GetProtectFromNukes(GetEffectiveBuildings())` (`CityData.cpp:3030`).
  - `CityData::HasForceField()` calls `buildingutil_GetForceField(GetEffectiveBuildings())` (`CityData.cpp:3043-3045`).
  - Special-agent success chances subtract `ProtectFromBioAgents`, `ProtectFromNanoVirus`, and `PreventSlavery` via `GetEffectiveBuildings()` (`CityData.cpp:3133-3140`).
  - Conversion protection returns `buildingutil_GetPreventConversion(GetEffectiveBuildings())` (`CityData.cpp:3456`).
  - Slavery protection returns `buildingutil_GetPreventSlavery(GetEffectiveBuildings())` (`CityData.cpp:4527`).
- Retail helper semantics are max/best for percentage shields and boolean for force-field/nuke shields in `buildingutil.cpp:183-203`, `228-253`, and `355-361`.

## Server coverage

- Building protection fields import to `EffectProtectFromNukes`, `EffectProtectFromBioAgents`, `EffectProtectFromNanoVirus`, `EffectPreventConversion`, `EffectPreventSlavery`, and `EffectForceField`.
- Nuclear, special-agent, conversion, and slavery handlers use `cityEffects`, which now includes wonder-granted effective-building effects from Task 107 while preserving direct building behavior.
- Added `TestRetailEffectiveBuildingProtectionBlocksNuclearAttack` to lock the effective-building force-field path.

## Validation

- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-110-target-green.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-110-green.txt`
