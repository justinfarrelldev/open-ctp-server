# CTP2 Full Retail Task 157: social slider happiness

## Implementation
- Nations now persist optional workday, wages, and rations slider values.
- Retail government workday/wages/rations expectations and coefficients become active nation effects when changing government.
- City happiness applies the retail source formula: `(slider - expectation) * positive/negative coefficient`, using centi-scaled integer coefficients for deterministic state.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailSocialSliderGovernmentCoefficientsAffectHappiness -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Public commands/projections for changing workday/wages/rations and economy-side effects remain future parity work.
