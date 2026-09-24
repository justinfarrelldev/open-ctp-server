# CTP2 Full Retail Task 119: city conversion chance

## Implementation
- `special_convert_city` effect values now act as deterministic percent success chances.
- Legacy/synthetic conversion effects with no positive value still behave as 100% to preserve existing bounded fixtures.
- Failed chance rolls spend the unit move and emit `city.conversion_failed` without changing ownership.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailConvertCityHonorsProtectionAndDisclosure -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Exact CTP2 capture/escape/diplomacy aftermath for failed conversions remains future work.
