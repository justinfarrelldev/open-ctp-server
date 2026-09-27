# CTP2 Full Retail Tasks 292-295: indulgence-sales special execution

## Implementation
- Added `indulgence-sales` as an executable special-action command for retail `IndulgenceSales` units.
- Execution spends special movement, transfers the retail default unconverted indulgence gold amount (3) from target nation to actor nation, and applies the retail default one-turn happiness bonus (+2) to the target city.
- Event payloads remain aggregate/disclosed and do not leak canonical city identifiers.
- Safe projections, HTTP routes, and AI scoring expose/use the action only from owned capable units.

## Source notes
- Retail `ArmyData::IndulgenceSale` transfers `UNCONVERTED_INDULGENCE_GOLD` and applies a 1-turn `UNCONVERTED_INDULGENCE_HAPPINESS` timer for unconverted cities. Converted/other-faith branches require full conversion ownership modeling and remain future parity work.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailIndulgenceSalesTransferGoldAndTimedHappiness -count=1`
- `go test ./internal/gameplay/visibility -run TestTask12ProjectionIncludesRulesAndLegalActionMetadata -count=1`
- `go test ./internal/gameplay/http -count=1`
- `go test ./internal/gameplay/ai -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
