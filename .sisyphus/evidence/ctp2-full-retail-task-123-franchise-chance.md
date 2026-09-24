# CTP2 Full Retail Task 123: franchise chance

## Implementation
- `special_create_franchise` positive effect values now act as deterministic percent success chances.
- Failed chance rolls spend movement, do not add a franchise marker, and emit disclosure-safe `city.franchise_failed`.
- Zero-value legacy/synthetic franchise effects remain 100% success for existing fixtures.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailCreateFranchiseMarksCityOnce -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail franchise income transfer, duration, lawsuit court outcomes, and diplomacy penalties remain future work.
