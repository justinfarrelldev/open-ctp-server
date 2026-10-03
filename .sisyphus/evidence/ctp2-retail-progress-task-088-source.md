# Task 088 source notes — building structural marker closure

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/BuildingRecord.h:152,165,255,262` exposes `Capitol`, `DoubleTelevangelists`, `Television`, and `Brokerage` building flags.
- `ctp2_source_code/ctp2_code/gs/gameobj/buildingutil.cpp:104-108` scans built improvements for `GetCapitol()`.
- `ctp2_source_code/ctp2_code/gs/gameobj/buildingutil.cpp:154-158` scans built improvements for `GetDoubleTelevangelists()`.
- `ctp2_source_code/ctp2_code/gs/gameobj/buildingutil.cpp:258-262` scans built improvements for `GetTelevision()`.
- `ctp2_source_code/ctp2_code/gs/gameobj/buildingutil.cpp:329-333` scans built improvements for `GetBrokerage()` and `buildingutil.cpp:392-405` applies wonder `IncreaseBrokerages` to brokerage commerce.
- `ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:2350-2361` uses `wonderutil_GetGoldPerTelevision` plus `buildingutil_GetTelevision` to award television wonder gold.
- `ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:3081-3088` treats `Capitol` buildings specially during city/building cleanup.

Server closure:

- `Capitol` is imported and executed for capture/bombard protection and empire-capital effects.
- `DoubleTelevangelists` is imported and executed for televangelist conversion percent gating.
- `Television` is imported and executed for city gold-per-population effects and wonder `GoldPerTelevision` income, including effective-building coverage.
- `Brokerage` is imported and executed with wonder `IncreaseBrokerages` commerce amplification.

Validation:

- `.sisyphus/evidence/ctp2-retail-progress-task-088-validation.txt` covers the existing targeted structural marker tests.
