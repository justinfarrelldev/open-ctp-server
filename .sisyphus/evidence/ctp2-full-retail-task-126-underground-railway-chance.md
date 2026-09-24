# CTP2 Full Retail Task 126: underground-railway chance

## Implementation
- `special_underground_railway` positive effect values now act as deterministic percent success chances.
- Failed chance rolls spend movement, do not reduce target population, do not add freed-population markers, and emit disclosure-safe `city.underground_railway_failed`.
- Zero-value legacy/synthetic effects remain 100% success for existing fixtures.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailUndergroundRailwayFreesPopulation -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail death chance and slave-object transfer details remain future work.
