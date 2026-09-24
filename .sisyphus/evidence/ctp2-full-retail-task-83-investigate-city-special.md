# CTP2 Full Retail Task 83: investigate city special action

## Source references
- Retail unit field `InvestigateCity` marks units able to inspect enemy cities.
- Investigation reveals city details to the acting player, not globally.

## Implementation
- Added `investigate-city` command and disclosure-safe event.
- Capable adjacent owned units spend movement and write/update a private city report on the actor nation.
- City reports include city owner, tile, turn, population, building IDs, and city effects.
- Player projections include only the viewing nation's reports; public events avoid target identifiers.
- State clone/normalization/validation now handle city reports deterministically.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailInvestigateCityStoresPrivateCityReport -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail report fields, report aging, UI redaction, and intelligence sharing still need parity review.
