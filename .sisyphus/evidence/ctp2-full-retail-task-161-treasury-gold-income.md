# CTP2 Full Retail Task 161: treasury gold income

## Implementation
- Nations now persist a `gold` treasury separate from public works.
- Turn advancement accrues the non-science portion of each owned city's net gold/trade yield into the nation treasury, honoring active science percent and government max-science caps.
- Owner diplomacy projections include only that owner's treasury gold.
- State validation rejects negative treasury gold.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailTreasuryAccruesNonScienceCityGold -count=1`
- `go test ./internal/gameplay/visibility -run TestTreasuryGoldProjectsOnlyToOwner -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Retail gold stats, maintenance bankruptcy, and rush-buy spending remain future economy parity work.
