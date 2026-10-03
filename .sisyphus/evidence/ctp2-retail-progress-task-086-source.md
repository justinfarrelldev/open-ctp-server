# Task 086 source notes — building CantBuildOnLand placement gate

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/BuildingRecord.cpp:676-680` parses both `CantBuildInSea` and `CantBuildOnLand` as building flags.
- `ctp2_source_code/ctp2_code/gs/newdb/BuildingRecord.h:259-260` exposes `GetCantBuildInSea()` and `GetCantBuildOnLand()`.
- `ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:4276-4285` checks the city tile: if the city is water and `CantBuildInSea` is set, production is rejected; otherwise, if it is land and `CantBuildOnLand` is set, production is rejected.

Implementation notes:

- Runtime production now carries `CantBuildOnLand` alongside the existing `CantBuildInSea` and `CoastalOnly` placement flags.
- Retail projection imports `CantBuildOnLand` from building flags.
- Engine production validation, public visibility/legal-action projection, and AI production placement selection now all reject land-city production when `CantBuildOnLand` is set and allow it in sea cities.

Validation:

- RED evidence: `.sisyphus/evidence/ctp2-retail-progress-task-086-red.txt`.
- GREEN targeted evidence: `.sisyphus/evidence/ctp2-retail-progress-task-086-green-targeted.txt`.
- GREEN package evidence: `.sisyphus/evidence/ctp2-retail-progress-task-086-green-packages.txt`.
- Freeze evidence: `.sisyphus/evidence/ctp2-retail-progress-task-086-freeze.txt`.
- Full suite evidence: `.sisyphus/evidence/ctp2-retail-progress-task-086-green-full.txt`.
