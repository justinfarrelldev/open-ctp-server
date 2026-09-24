# CTP2 Full Retail Task 113: safe trade-route projection

## Implementation
- Added viewer trade-route projections for canonical trade routes.
- Routes are projected only when the viewer owns the source city and both endpoint city tiles are currently visible.
- Hidden destination cities/routes remain absent from projections until normal visibility (for example global radar) reveals them.

## Tests
- `go test ./internal/gameplay/visibility -run TestRetailTradeRouteProjectionDoesNotLeakHiddenDestinations -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Trade-route creation commands, route path projections, pillage/piracy visibility, and SSE event coverage remain future work.
