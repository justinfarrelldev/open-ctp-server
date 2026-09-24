# CTP2 Full Retail Task 91: injoin special action

## Source references
- Retail unit field `CanInjoin` marks units able to legally enjoin enemy cities.
- Injunction affects city operations rather than ordinary combat.

## Implementation
- Added `injoin` command and disclosure-safe event.
- Capable adjacent owned units spend movement and apply a city-local `injunction` effect.
- Cities with an injunction do not advance their production queue while the effect remains.
- State validation accepts the injunction effect marker.
- Owner legal-action projections expose visible adjacent injunction targets for capable units not already enjoined.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailInjoinCityBlocksProduction -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail injunction duration, court outcomes, counter-actions, and diplomacy remain to be implemented.
