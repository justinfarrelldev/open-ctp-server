# CTP2 Full Retail Task 95: food vat consumption effect

## Source references
- Retail building marker `FoodVat` is a city food/growth effect.
- Food vat behavior affects starvation/growth by reducing effective food pressure.

## Implementation
- City food advancement now applies the maximum city-local `food_vat_percent` effect as a food-consumption reduction.
- Reduction is clamped at 100% and integrates with existing starvation protection and growth thresholds.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailFoodVatReducesFoodConsumption|TestRetailStarvationProtectionDelaysPopulationLoss' -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Exact retail FoodVat percentages/source data semantics and interactions with specialists/granaries remain to be audited.
