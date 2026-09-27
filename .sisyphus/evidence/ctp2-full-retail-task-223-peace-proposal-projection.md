# CTP2 Full Retail Task 223: project peace proposals while at war

## Implementation
- Safe diplomacy projection now exposes `propose-agreement` actions with `kind:"peace"` for contacted relations whose visible regard is `hotwar` and that lack a pending peace proposal.
- War relations no longer project union/trade proposals in the same pass; peace is the public route out of war.

## Tests
- `go test ./internal/gameplay/visibility -run 'TestDiplomacyProjects(PeaceProposalAtWar|TradeProposalBeforeForeignTradeRoutes)' -count=1`
- `go test ./internal/gameplay/visibility -count=1`
- `go test ./internal/gameplay/ai -count=1`
