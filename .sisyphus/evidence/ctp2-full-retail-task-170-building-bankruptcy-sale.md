# CTP2 Full Retail Task 170: building bankruptcy sale

## Source evidence
- Retail gold deficit paths can force loss/sale of city improvements when a player cannot cover city maintenance costs.
- Server state must never carry negative treasury balances.

## Implementation
- When treasury advancement would drive a nation below zero, the engine removes highest-upkeep owned buildings deterministically until the shortfall is covered or no upkeep buildings remain.
- Tie-breaking is deterministic by city/building ID.
- Treasury is then clamped at zero.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailTreasuryShortfallSellsHighestUpkeepBuilding -count=1`
- `go test ./internal/gameplay/engine -count=1`
