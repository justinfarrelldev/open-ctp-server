# CTP2 Full Retail Task 221: war cancels combatant trade routes

## Source evidence
- Retail diplomacy/trade route flow couples trade with diplomatic state; hostile transitions invalidate peaceful trade.
- The server already cleared trade/treaty relation state on war, but left existing route records intact.

## Implementation
- Declaring war now removes canonical trade routes whose source/destination cities are owned by the two new combatants.
- Unrelated third-party routes are preserved.

## Tests
- `go test ./internal/gameplay/engine -run TestSetWarRemovesTradeRoutesBetweenCombatants -count=1`
- `go test ./internal/gameplay/engine -count=1`
