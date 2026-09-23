# CTP2 full-retail parity task 19 retail terrain yields evidence

Scope: use projected retail terrain EnvBase yields for city yield calculation when a pinned runtime terrain table provides them.

Original source/data checked:
- Retail `terrain.txt` `TERRAIN_PLAINS` has `EnvBase Food 10`, `Shield 10`, `Gold 5`, and `Movement 100`.
- Terrain yield values in retail data are stored as authoritative scalar values and are now projected into runtime terrains.

Implemented server-compatible slice:
- `RuntimeTerrain` now carries `food`, `production`, and `gold` in addition to movement cost/domains.
- Retail runtime projection maps `EnvBase.Food`, `EnvBase.Shield`, and `EnvBase.Gold`.
- Synthetic terrain records can optionally provide `Food`, `Production`, and `Gold` without changing existing synthetic fallbacks.
- Round recalculation uses runtime terrain yields when present, while old direct `terrainYield` calls keep fallback behavior.
- City production accumulation now also uses runtime terrain production when present.
- Completed terrain-improvement overlay yields are still added on top of runtime base terrain yields.

Validation commands:
- `go test ./internal/gameplay/engine -run 'RetailRuntimeTerrainYields(AccumulateProduction|DriveCityYield)' -count=1`
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`
