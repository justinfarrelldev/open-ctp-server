# Task 060 source check: television wonder gold uses effective buildings

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:2350-2361` computes `GoldPerTelevision` wonder income by scanning foreign cities and checking `buildingutil_GetTelevision(city.CD()->GetEffectiveBuildings())`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:5717-5722` defines effective buildings as local built improvements plus owner `GetWonderBuildings()` from building-everywhere wonders.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/buildingutil.cpp:258-263` resolves the `Television` building flag from the effective-building bitset.

Implementation notes:

- `cityHasTelevision` now checks `cityLocalEffects`, so active wonder-granted effective building effects participate in `GoldPerTelevision` payouts instead of only directly built city buildings.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-060-red.txt`
- GREEN engine/package: `.sisyphus/evidence/ctp2-retail-progress-task-060-green-engine.txt`, `.sisyphus/evidence/ctp2-retail-progress-task-060-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-060-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-060-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-060-diff-check.txt`
