# CTP2 full-retail parity task 11 retail building yield effects evidence

Scope: execute a first retail building-effect slice by projecting production/commerce percentage bonuses into runtime production effects and applying them after building completion.

Original source/data checked:
- `ctp2_code/gs/newdb/building.cdb` defines `ProductionPercent` and `CommercePercent` as building float fields.
- `ctp2_code/gs/gameobj/buildingutil.cpp::buildingutil_GetProductionPercent` aggregates built-improvement production percentage bonuses.
- `ctp2_code/gs/gameobj/CityData.cpp` applies the aggregated production percentage to gross city production using `ceil(gross_production * prodBonus)`.
- Retail data examples verified: `IMPROVE_FACTORY` has `ProductionPercent 0.15`; `IMPROVE_BAZAAR` has `CommercePercent 0.1`.

Implemented server-compatible slice:
- Runtime production records can now carry executable effects.
- Retail projection maps `ProductionPercent` to `yield_production_percent` and `CommercePercent` to `yield_gold_percent` using percentage points.
- Engine queueing copies pinned runtime production effects into completed buildings/wonders instead of trusting client-provided effects.
- City yield calculation applies positive percentage bonuses with ceiling semantics before upkeep.

Validation commands:
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run RuntimeProductionEffectsApplyToCompletedBuildingYields -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
