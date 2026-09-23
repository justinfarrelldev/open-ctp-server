# CTP2 full-retail parity task 10 retail production costs evidence

Scope: project and enforce retail building/wonder production costs and advance unlocks in the runtime queue path.

Original source/data checked:
- `ctp2_code/gs/gameobj/bldque.cpp` loads queue entries with costs from `UnitRecord::GetShieldCost`, `BuildingRecord::GetProductionCost`, and `WonderRecord::GetProductionCost`.
- `ctp2_code/gs/gameobj/CityData.cpp` build insertion paths likewise use unit shield cost and building/wonder production cost from the DB records.
- Retail data examples verified: `IMPROVE_GRANARY` has `EnableAdvance ADVANCE_AGRICULTURE`, `ProductionCost 300`, `Upkeep 1`; `WONDER_HAGIA_SOPHIA` has `EnableAdvance ADVANCE_THEOLOGY`, `ProductionCost 3240`.

Implemented server-compatible slice:
- `RuntimeProduction` now carries cost, upkeep, prerequisite, and obsolescence metadata.
- Retail runtime projection fills production cost/unlock data for units, buildings, and wonders.
- Synthetic record runtime compile can also populate building/wonder production defaults from supported tags.
- Engine queueing normalizes building/wonder/unit cost and gates production prerequisites from pinned runtime data before accepting the item.

Validation commands:
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run RuntimeBuildingAndWonderProductionUsesPinnedCostAndPrerequisite -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
