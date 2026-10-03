# Task 109 source evidence: building crime/pollution effects

## Retail source checks

- Building pollution fields are parsed in `BuildingRecord` (`PopulationPollutionPercent`, `ProductionPollutionPercent`; `ctp2_source_code/ctp2_code/gs/newdb/BuildingRecord.cpp:276-279`).
- Retail city pollution iterates `GetEffectiveBuildings()` and accumulates building pollution fields (`CityData.cpp:1077-1088`), so wonder-granted effective buildings can participate in the same helper path.
- Retail crime reduction also uses effective buildings: `CityData::GetImprovementCrimeMod()` returns `buildingutil_GetLowerCrime(GetEffectiveBuildings())` (`CityData.cpp:2808-2811`), and `buildingutil_GetLowerCrime` sums `GetLowerCrime` over built improvements (`buildingutil.cpp:172-179`).
- Retail `EliminateNukes` destroys actual nuclear-plant improvements and then removes illegal queue items (`CityData.cpp:4869-4880`).
- Released `buildings.txt` contains `LowerCrime`, `PopulationPollutionPercent`, `ProductionPollutionPercent`, and `NuclearPlant`, but no `PollutionAmount` records; flat building pollution amount is therefore source-classified as a parsed source capability absent from released data for current retail runtime coverage.

## Server coverage

- `LowerCrime` imports as `crime_percent` and drives city crime calculations through `cityEffects`, now including effective-building propagation from Task 107.
- `PopulationPollutionPercent` and `ProductionPollutionPercent` import as pollution effects and drive `cityPollution`; advance pollution modifiers and existing pollution tests cover stacking/clamping.
- `NuclearPlant` production/building removal under global `EliminateNukes` is already covered by the nuclear-plant/eliminate-nuke task tests.
- Because released data has no building `PollutionAmount`, no additional runtime execution is required for literal CTP2 retail data.

## Validation

- Existing focused tests: `go test ./internal/gameplay/engine -run 'TestRetailBuildingCrimeEffectsStackWithWonderCrimeEffects|TestRetailPollution|TestRetailNuclearPlant|TestRetailEliminateNukes' -count=1`
- Package/ledger validation: `.sisyphus/evidence/ctp2-retail-progress-task-109-green.txt`, `.sisyphus/evidence/ctp2-retail-progress-task-109-field-ledger-green.txt`
