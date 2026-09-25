# CTP2 Full Retail Task 165: readiness support production tax

## Source evidence
- `Readiness.cpp` computes support from unit shield hunger, readiness coefficient, special-forces war coefficient, support coefficient, and readiness-cost-reduction wonders.
- `Player.cpp` subtracts readiness cost from city production using a federal production percentage after franchise production offsets support cost.
- `CityData::PayFederalProduction` pays `ceil(city_production * percent)` for all but the final city, which pays the remainder.

## Implementation
- End-of-round city yield recalculation now applies a deterministic nation-wide readiness support production tax.
- Support cost ignores `NeedsNoSupport`, uses war readiness by default, honors government support/readiness coefficients, special-forces war coefficients, and readiness-cost-reduction wonders.
- Franchise production offsets support cost before city production is reduced.
- Over-budget unit disbanding remains a future parity step; current tax is capped at available city production for state safety.

## Tests
- `go test ./internal/gameplay/engine -run TestReadinessSupportReducesCityProduction -count=1`
- `go test ./internal/gameplay/engine -count=1`
