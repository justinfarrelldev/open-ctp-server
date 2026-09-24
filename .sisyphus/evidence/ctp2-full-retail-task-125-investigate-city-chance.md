# CTP2 Full Retail Task 125: investigate-city chance

## Implementation
- `special_investigate_city` positive effect values now act as deterministic percent success chances.
- Failed chance rolls spend movement, do not create/update private city reports, and emit disclosure-safe `city.investigation_failed`.
- Zero-value legacy/synthetic effects remain 100% success for existing fixtures.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailInvestigateCityStoresPrivateCityReport -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail death/capture/escape chances for investigating units remain future work.
