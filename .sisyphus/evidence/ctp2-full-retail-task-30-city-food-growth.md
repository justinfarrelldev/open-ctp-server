# CTP2 full-retail parity task 30 city food growth evidence

Scope: begin executing the CTP2 city food-growth pipeline with canonical city population state.

Source notes:
- `CityData::ProcessFood`, `EatFood`, `CalculateGrowthRate`, and `GrowOrStarve` compute food produced, food required per citizen, surplus/deficit, and population growth/starvation through partial population progress.

Implemented bounded model:
- Cities now persist `StoredFood` alongside `Population`.
- Recalculation applies food surplus after city yield computation.
- Population consumes food at 2 per population per recalculation, accumulates surplus in `StoredFood`, grows when reaching `100 * population`, and starves down when stored food is exhausted.
- Owned viewer projections expose `population`/`stored_food`; enemy city economy remains hidden.

Validation commands:
- `go test ./internal/gameplay/engine -run RetailCityFoodSurplusGrowsPopulation -count=1`
- `go test ./internal/gameplay/visibility -run Task7ProjectionExposesOwnedCityEconomyOnly -count=1`
