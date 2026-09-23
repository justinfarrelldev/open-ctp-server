# CTP2 full-retail parity task 28 production API/AI building choices evidence

Scope: expose runtime building production safely through viewer projections/legal actions and let projected-rule AI choose yield buildings before units.

Implemented:
- Viewer `RulesProjection` now includes safe `RuntimeProduction` records.
- Legal production actions now include unlocked runtime buildings/wonders from public projection data, including pinned costs/prereqs/effects.
- AI production planning scores unlocked city buildings with beneficial effects before falling back to unit production.
- Authoritative queue validation rejects an already-built city building, preserving the one-building-per-city invariant before AI uses building production more often.

Validation commands:
- `go test ./internal/gameplay/visibility -run ProductionActionsIncludeUnlockedRuntimeBuildings -count=1`
- `go test ./internal/gameplay/ai -run 'PlannerQueuesProjectedRetailYieldBuildingBeforeUnits|PlannerQueuesProjectedRetailLikeUnitProduction' -count=1`
- `go test ./internal/gameplay/engine -run RuntimeBuildingProductionRejectsAlreadyBuiltCityImprovement -count=1`
