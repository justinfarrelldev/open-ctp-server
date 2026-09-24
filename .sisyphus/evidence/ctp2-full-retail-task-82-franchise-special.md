# CTP2 Full Retail Task 82: create franchise special action

## Source references
- Retail unit field `CreateFranchise` marks units able to franchise enemy cities.
- Retail buildings/wonders include brokerage/franchise-related markers that depend on city franchise state.

## Implementation
- Added `create-franchise` command and disclosure-safe event.
- Capable adjacent owned units spend movement and mark the enemy city with a player-scoped franchise effect.
- Duplicate franchises by the same owner are rejected.
- State validation now accepts the franchise city effect.
- Owner legal-action projections expose create-franchise only for visible adjacent enemy cities without an existing franchise from that player.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailCreateFranchiseMarksCityOnce -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Full retail franchise income, brokerage multipliers, diplomatic effects, and removal/expiration rules remain to be implemented.
