# CTP2 full-retail parity task 07 runtime production costs evidence

Scope: make unit production queueing use pinned runtime unit costs instead of trusting client-provided costs when a runtime definition is available.

Original source checked:
- `ctp2_code/gs/gameobj/CityData.cpp` production selection path: when city production is changed to a unit, the build queue item cost is set from `UnitRecord::GetShieldCost()`.
- `ctp2_code/gs/gameobj/bldque.cpp::BuildQueue::BuildFrontUnit`: completion compares the build queue node cost against stored city production before creating the unit.

Implemented server-compatible slice:
- Queueing a runtime-known unit normalizes `ProductionItem.Cost` to `RuntimeUnitDefinition.ProductionCost` when positive.
- Runtime unit upkeep/prerequisite/obsolete defaults are also copied into the queued item when present.
- Synthetic/no-cost runtime units preserve existing behavior.

Validation commands:
- `go test ./internal/gameplay/engine -run RuntimeUnitProductionUsesPinnedShieldCost -count=1`
- `go test ./internal/gameplay/engine -count=1`
