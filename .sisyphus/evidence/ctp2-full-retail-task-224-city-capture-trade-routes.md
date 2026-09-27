# CTP2 Full Retail Task 224: city capture cancels affected trade routes

## Implementation
- Capturing a city now removes canonical trade routes whose source or destination is the captured city.
- Unrelated trade routes are preserved.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailCityCaptureRemovesTradeRoutesTouchingCapturedCity -count=1`
- `go test ./internal/gameplay/engine -count=1`
