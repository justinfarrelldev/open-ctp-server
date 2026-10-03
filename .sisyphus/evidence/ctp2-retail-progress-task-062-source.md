# Task 062 source check: captured capitals lose Capitol buildings

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:1958-1962` calls `m_city_data->DestroyCapitol()` while changing city owner during `ResetCityOwner`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:2662-2675` implements `CityData::DestroyCapitol` by clearing built improvements whose building util marks them as designating the capitol, then recalculating defensive bonus from effective buildings.

Implementation notes:

- `captureCityAt` now removes captured city buildings carrying the `capitol` effect before changing owner.
- Non-capitol buildings are preserved.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-062-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-062-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-062-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-062-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-062-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-062-diff-check.txt`
