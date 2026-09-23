# CTP2 full-retail parity task 27 terrain EnvCity yields evidence

Scope: preserve and execute retail terrain `EnvCity` yield modifiers for city-center yields.

Original source/data checked:
- Retail `terrain.txt` defines `EnvBase` and `EnvCity` blocks for terrains. Example `TERRAIN_PLAINS` has `EnvBase` food/shield/gold `10/10/5`, while `EnvCity` is `10/15/15`.

Implemented:
- Typed retail terrain catalog now preserves `EnvCity` modifiers.
- Runtime terrain projection carries city-center food/production/gold fields separately from base worked-tile yields.
- Engine city yield and production accumulation use runtime city-center yield for the city tile, while worked tiles continue using base terrain yield.

Limitations pending full parity:
- `EnvRiver`, dead terrain yields, resource goods, and transform gates are still pending.

Validation commands:
- `go test ./internal/gameplay/rules -run 'LoadRetailRulesFromDirectoryPreservesTypedRetailValues|ProjectRetailRuntimeMapsSelectedRetailValues' -count=1`
- `go test ./internal/gameplay/engine -run RetailRuntimeTerrainYieldsDriveCityYield -count=1`
