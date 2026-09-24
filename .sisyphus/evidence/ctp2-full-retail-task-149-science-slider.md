# CTP2 Full Retail Task 149: nation science slider

## Implementation
- Nations can now persist an explicit `science_percent` pointer, allowing both unset/default (100%) and explicit zero science.
- Research generation applies the nation science percent to the city gold/trade research portion, then caps it by any active government `MaxScienceRate`.
- Flat city/nation research bonuses remain uncapped, preserving existing retail max-science behavior.
- Owner diplomacy projections include the safe science-percent value when explicitly set.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailResearchRateUsesNationSciencePercent -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Public commands/UI for changing tax/luxury/science sliders remain future government/economy parity work.
