# CTP2 Full Retail Tasks 280-283: settle city type and city growth coefficient execution

## Implementation
- Produced units now preserve `SettleCityTypeID` from runtime unit definitions.
- Found-city execution transfers a settler unit's `settle_city_type_id` into canonical `City.TypeID`.
- City food growth now uses the founded city type's runtime `CityGrowthCoefficientPercent`, matching the retail model where city unit records provide the growth coefficient (`UNIT_CITY=1`, `UNIT_OCEAN_CITY/UNIT_SPACE_CITY=0.5`).
- Owned-unit projections expose `settle_city_type_id`; visible enemy unit projections continue to suppress hidden unit capabilities.

## Source notes
- Retail `CityData::CityGrowthCoefficient()` multiplies the global city-growth coefficient by the city unit DB record's `CityGrowthCoefficient`.
- Retail settlers provide `SettleCityType` to determine the city unit record used by the founded city.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetail(SettlerEffectsSeedFoundedCitySizeBuildingsAndCityType|RuntimeSettleCityTypeAndGrowthCoefficient)' -count=1`
- `go test ./internal/gameplay/visibility -run TestTask5ProjectionExposesOwnedUnitCapabilitiesOnly -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
