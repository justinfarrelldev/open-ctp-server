# CTP2 Full Retail Task 178: trade route creation command

## Source evidence
- Retail `Player::CreateTradeRoute` validates source/destination cities, route type, and duplicate/capacity constraints before inserting a route into the trade pool.
- Retail `TradePool::Create` links the created route to source and destination city state.

## Implementation
- Added `create-trade-route` command and `trade_route.created` event.
- Validates owned source city, existing destination city, distinct source/destination, stable unique route IDs, and duplicate source/destination routes.
- Creates canonical `TradeRoute` entries with deterministic IDs and a conservative `CrossesWater` marker when the simple route corridor includes water.

## Tests
- `go test ./internal/gameplay/engine -run TestCreateTradeRoute -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail route path/cost, commodities/resources, route capacity, piracy, and treaty limitations remain future trade parity work.
