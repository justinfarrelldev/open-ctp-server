# CTP2 Full Retail Task 163: unit upkeep debits treasury

## Implementation
- Unit gold upkeep now debits the nation treasury during turn economy advancement instead of reducing science output directly.
- Treasury debits are clamped at zero to preserve safe state invariants.
- Research generation now uses the science portion of city gold without subtracting unit upkeep, matching the separation between gold maintenance and science accounting.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailUnitUpkeepDebitsTreasuryNotResearch -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail bankruptcy disband/sell behavior and detailed gold-stat accounting remain future economy parity work.
