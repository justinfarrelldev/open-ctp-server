# CTP2 Full Retail Tasks 328-331: city good radius inventory

## Source notes
- Retail `CityData::GetGoodCountInRadius` counts goods in a city radius and retail unit `NeedsCityGood` gates production on those goods.

## Implementation
- Canonical tiles can now preserve a `good_id` resource marker.
- Founded cities collect unique goods from their city tile and adjacent radius into `City.Goods`.
- City goods refresh during city recalculation so terrain/good changes can affect later production legality.
- `NeedsCityGood` production gates now work from map-derived city goods as well as explicitly seeded city goods.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRuntimeFoundCityCollectsAdjacentGoodsForNeedsCityGoodProduction|TestRuntimeProductionQueueRejectsUnitsNeedingAbsentCityGood' -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
