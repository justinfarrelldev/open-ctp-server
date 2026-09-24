# CTP2 Full Retail Task 160: pollution unhappiness

## Implementation
- Government `PollutionUnhappyCoef` becomes an active nation effect when changing government.
- City happiness subtracts a deterministic pollution unhappiness penalty from the city's tile pollution using centi-scaled coefficients.
- The penalty follows the retail formula shape: `-pollution_unhappy_coef * city_pollution`.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailPollutionUnhappyCoefReducesHappiness -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Wonders that explicitly remove pollution unhappiness and exact floating-point rounding remain future parity work.
