# CTP2 full-retail parity task 24 building FoodPercent evidence

Scope: project retail building `FoodPercent` fields into runtime effects and apply them to authoritative city food yield.

Original source/data checked:
- Retail `buildings.txt` includes `FoodPercent 0.1` for `IMPROVE_GRANARY` and `FoodPercent 0.15` for `IMPROVE_FOOD_SILO`.

Implemented:
- Runtime building projection now maps `FoodPercent` to `yield_food_percent`.
- Engine city yield calculation now applies `yield_food_percent` with the same rounded-up percentage helper used by production/gold percentage effects.
- Effect validation accepts `yield_food_percent` so runtime-pinned building effects can be queued and completed safely.

Limitations pending full parity:
- Starvation protection and population growth mechanics are still pending; this only wires the food yield modifier.

Validation commands:
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run RetailRuntimeBuildingFoodPercentDrivesCityYield -count=1`
