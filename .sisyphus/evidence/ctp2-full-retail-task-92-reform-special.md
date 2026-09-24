# CTP2 Full Retail Task 92: reform special action

## Source references
- Retail unit field `CanReform` marks units able to reform cities.
- Reform is modeled as a special city operation that can counter legal/civic disruption.

## Implementation
- Added `reform` command and disclosure-safe event.
- Capable adjacent owned units spend movement to reform an enemy city.
- Reform removes city injunction effects, clears city crime, and increases city happiness by one.
- Owner legal-action projections expose visible adjacent reform targets for capable units.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailReformCityRemovesInjunctionAndCrime -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail reform success chances, ownership/revolt interactions, duration, and diplomacy remain to be implemented.
