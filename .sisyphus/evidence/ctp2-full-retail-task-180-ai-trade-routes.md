# CTP2 Full Retail Task 180: AI projected trade routes

## Implementation
- AI now consumes projected `create-trade-route` legal actions.
- The planner validates projected payload shape and copies the payload without deriving hidden destinations from canonical state.
- Trade route candidates are scored conservatively below specialist/readiness/economy choices.

## Tests
- `go test ./internal/gameplay/ai -run TestPlannerUsesProjectedTradeRouteAction -count=1`
- `go test ./internal/gameplay/ai -count=1`
