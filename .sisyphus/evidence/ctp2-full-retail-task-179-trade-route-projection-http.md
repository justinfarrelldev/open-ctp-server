# CTP2 Full Retail Task 179: trade route projection and HTTP route

## Implementation
- Owner-safe legal actions now project `create-trade-route` for visible source/destination city pairs without existing visible routes.
- HTTP command routing accepts `create-trade-route` with `source_city_id` and `destination_city_id`.
- Existing safe route projection still hides routes whose endpoints are not visible to the viewer.

## Tests
- `go test ./internal/gameplay/visibility -run TestVisibleCitiesProjectCreateTradeRouteAction -count=1`
- `go test ./internal/gameplay/http -count=1`
- `go test ./internal/gameplay/visibility -count=1`
