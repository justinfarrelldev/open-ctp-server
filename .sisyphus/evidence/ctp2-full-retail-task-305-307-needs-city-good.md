# CTP2 Full Retail Tasks 305-307: NeedsCityGood production gating

## Source notes
- Retail data includes `NeedsCityGood` on `UNIT_COFFEE_PLANT` (`Units.txt`).
- Retail source models city goods in-radius through `CityData::GetGoodCountInRadius`, which counts matching goods around the city.

## Implementation
- Added canonical/persisted `City.Goods` as the server-side city-good inventory abstraction.
- Unit production rejects runtime units whose `needs_city_good_id` is absent from the producing city.
- Owned city projections expose city goods, and production action projection suppresses city-good-gated units unless the visible owned city has the required good.

## Tests
- `go test ./internal/gameplay/engine -run TestRuntimeProductionQueueRejectsUnitsNeedingAbsentCityGood -count=1`
- `go test ./internal/gameplay/visibility -run TestProductionActionsHideUnitsNeedingAbsentCityGood -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
