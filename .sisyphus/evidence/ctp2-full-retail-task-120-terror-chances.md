# CTP2 Full Retail Task 120: terror special-action chances

## Implementation
- Bio terror, nano terror, and plague now honor positive special-effect values as deterministic percent success chances.
- Existing zero-value synthetic effects remain 100% success for backwards-compatible bounded fixtures.
- Chance failures spend movement and emit the disclosure-safe `city.terror_failed` event.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailBioNanoAndPlagueSpecialsAffectCityPopulation -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Exact retail infection timers, production penalties, and capture/death side effects remain future work.
