# CTP2 Full Retail Task 94: sue special action

## Source references
- Retail unit field `CanSue` marks units able to bring legal action against enemy cities/franchises.
- Lawsuit actions disrupt city output rather than resolving as combat.

## Implementation
- Added `sue` command and disclosure-safe event.
- Capable adjacent owned units spend movement to sue enemy cities.
- A lawsuit clears stored production and applies a city-local `lawsuit` marker; duplicate lawsuits are rejected.
- State validation accepts the lawsuit effect marker.
- Owner legal-action projections expose visible adjacent lawsuit targets for capable units not already under lawsuit.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailSueCityAppliesLawsuit -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail lawsuit targets, franchise damages, court resolution, diplomacy, and duration remain to be implemented.
