# CTP2 Full Retail Tasks 300-304: selected Const.txt runtime constants

## Implementation
- Imported selected retail `Const.txt` values into `RetailRules.Constants` and runtime projections:
  - `CITY_GROWTH_COEFFICIENT`
  - `UNCONVERTED_INDULGENCE_GOLD`
  - `UNCONVERTED_INDULGENCE_HAPPINESS`
- City food growth now uses the runtime base city-growth coefficient before applying city-unit `CityGrowthCoefficient` percentages.
- Indulgence-sales execution now uses runtime indulgence gold/happiness constants, with retail defaults preserved as fallbacks for synthetic runtimes.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetail(RuntimeSettleCityTypeAndGrowthCoefficient|IndulgenceSalesTransferGoldAndTimedHappiness)' -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
