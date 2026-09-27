# CTP2 Full Retail Tasks 199-200: trade routes require trade agreements

## Source evidence
- Retail trade routes are created through trade-offer/diplomacy paths (`TradeOfferData::Accept` creates the paired route after an accepted offer).
- War transitions in the server already clear trade/treaty state; foreign routes should not be created without a diplomatic trade/union relationship.

## Implementation
- Canonical `create-trade-route` now allows domestic routes, but foreign destinations require visible/contacted relation state with `Trade`, `trade`, or `union`, and no active war edge.
- Safe legal-action projection applies the same owner-visible gating and suppresses foreign route actions without a trade agreement.
- Existing duplicate and hidden-route protections remain intact.

## Tests
- `go test ./internal/gameplay/engine -run 'TestCreateTradeRoute(CommandAddsCanonicalRoute|RejectsForeignDestinationWithoutTradeAgreement)' -count=1`
- `go test ./internal/gameplay/visibility -run 'TestVisibleCities(ProjectCreateTradeRouteAction|SuppressForeignTradeRouteActionWithoutTradeAgreement)' -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
- `go test ./internal/gameplay/ai -count=1`
