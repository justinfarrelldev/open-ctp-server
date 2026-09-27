# CTP2 Full Retail Tasks 270-275: additional retail unit special field preservation

## Implementation
- Preserved additional retail `Units.txt` gameplay fields in runtime/canonical definitions:
  - `CauseUnhappiness` -> `special_cause_unhappiness`, timer, and amount effects.
  - `IndulgenceSales` -> `special_indulgence_sales` marker effect.
  - `InciteRevolution` -> chance, elite chance, and death chance effects.
  - `SlaveUprising` -> `special_slave_uprising` marker effect.
  - `NeedsCityGood` -> `needs_city_good_id` runtime unit field.
  - `SettleCityType` and `CityGrowthCoefficient` -> runtime unit city-type/growth fields.
- These values are preserved for deterministic runtime, AI/projection visibility, and future execution. Full execution for these special actions/city-type semantics remains separate work.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
