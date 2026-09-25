# CTP2 Full Retail Task 168: social policy economy effects

## Source evidence
- `PlayHap.cpp` converts unitless sliders to per-citizen values using `BASE_WORKDAY`, `UNIT_WORKDAY`, `BASE_WAGES`, `UNIT_WAGES`, `BASE_RATIONS`, and `UNIT_RATIONS`.
- `CityData::ComputeGrossProduction` multiplies collected production by workday-per-person.
- `CityData::PayWages` subtracts wages per citizen from city trade/gold.
- `CityData::GetFoodRequired` consumes rations per citizen from produced food.

## Implementation
- Workday now adjusts city production using retail constants (`1.0 - workday * 0.15`) with safe clamping.
- Wages now subtract per-citizen gold using retail base/unit values (`4 + wages`).
- Rations now adjust per-city food consumption using a scaled retail base/unit ratio compatible with the server's existing food scale.

## Tests
- `go test ./internal/gameplay/engine -run TestSocialPolicyChangesProductionGoldAndFoodConsumption -count=1`
- `go test ./internal/gameplay/engine -count=1`
