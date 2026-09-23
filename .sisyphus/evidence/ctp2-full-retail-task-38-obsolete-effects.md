# CTP2 full-retail parity task 38 advance-driven obsolete effects evidence

Scope: completed buildings/wonders with retail `ObsoleteAdvance` metadata must stop contributing effects when the relevant nation owns the obsoleting advance.

Implemented:
- Added player/nation-aware obsolescence helper that combines existing manual city obsolete markers with nation advances.
- City yield effects skip buildings obsoleted by the owning nation's advances.
- Research-rate effects skip obsoleted science buildings.
- Wonder effects now use the wonder owner's advances for obsolescence while preserving existing manual obsolete-marker behavior.

Validation commands:
- `go test ./internal/gameplay/engine -run 'ObsoleteRetailBuildingScience|RuntimeBuildingEffectsObsoleteWhenNationHasObsoleteAdvance|RetailResearchRateUsesCityScienceBuildingEffect|RuntimeWonderEffectsDriveOwnedCityYields' -count=1`
- `GOMAXPROCS=4 go test ./... -count=1`
- `git diff --check`
