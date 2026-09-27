# CTP2 Full Retail Task 222: project trade agreement proposals

## Implementation
- Safe diplomacy projection now exposes `propose-agreement` actions with `kind:"trade"` for contacted, non-war relations that do not already have trade/union and do not already have a pending trade proposal.
- This gives AI/clients a public, legal route to create the trade agreements required by foreign trade-route creation.

## Tests
- `go test ./internal/gameplay/visibility -run TestDiplomacyProjectsTradeProposalBeforeForeignTradeRoutes -count=1`
- `go test ./internal/gameplay/visibility -count=1`
- `go test ./internal/gameplay/ai -count=1`
