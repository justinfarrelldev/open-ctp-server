# Task 108 source evidence: building FoodVat/population-happiness closure

## Retail source checks

- Retail parses population/happiness building fields in `BuildingRecord`: `FoodVat`, `RaiseOvercrowdingLevel`, `RaiseMaxPopulation`, and `StarvationProtection` are database fields (`ctp2_source_code/ctp2_code/gs/newdb/BuildingRecord.cpp:231-268`).
- Retail uses actual city improvements for food-vat starvation-warning behavior:
  - `newturncount.cpp:436` and `TurnCnt.cpp:728` call `buildingutil_HaveFoodVat(unit->GetImprovements(), tmp)` before starvation warnings.
  - `buildingutil.cpp:296-306` reads the maximum `GetFoodVat` coefficient across built improvements.
- Retail uses effective buildings for the rest of the population/happiness helpers:
  - `CityData.cpp:1520` and `1527` use `GetEffectiveBuildings()` for overcrowding and max-population caps.
  - `CityData.cpp:5688-5696` adds `buildingutil_GetStarvationProtection(GetEffectiveBuildings())`.
  - `buildingutil.cpp:129-145` sums `HappyInc`, treats `IsReligious` as a no-op multiplier in released source, and applies `Cathedral` only through a separate `IncreaseCathedrals` wonder bonus.
- Released wonder data has no `IncreaseCathedrals` field (`rg IncreaseCathedrals ctp2_released_game/ctp2_data/default/gamedata/Wonder.txt` returns no records), so `Cathedral`/`IsReligious` are preserved markers with no additional released-data gameplay mutation beyond `HappyInc`.

## Server implementation

- The existing food-vat food-consumption model now reads actual non-obsolete building `food_vat_percent` effects (and existing city markers) instead of city markers only.
- Existing population/happiness coverage remains source-aligned for `HappyInc`, `RaiseMaxPopulation`, `RaiseOvercrowdingLevel`, `StarvationProtection`, `LowerPeaceMovement`, and `TemporaryFullHappiness`; effective building propagation from Task 107 now also reaches the city-local effect paths those mechanics use.
- `Cathedral` and `IsReligious` remain imported marker effects and are classified as no-extra-mutation for released retail data because the only source-side modifier path depends on absent `IncreaseCathedrals` data and `IsReligious` subtracts zero.

## Validation

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-108-red.txt`
- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-108-green-target.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-108-green.txt`
- Freeze RED/GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-108-freeze-red.txt`, `.sisyphus/evidence/ctp2-retail-progress-task-108-freeze-green.txt`
- Fullgame GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-108-fullgame-green.txt`
