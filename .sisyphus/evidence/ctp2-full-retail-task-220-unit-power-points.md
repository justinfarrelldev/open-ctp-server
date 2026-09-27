# CTP2 Full Retail Task 220: unit PowerPoints preservation

## Source evidence
- Retail `Units.txt` defines `PowerPoints` for units (for example `UNIT_AIRCRAFT_CARRIER` has `PowerPoints 3000`).
- The typed retail catalog already preserved the source value; runtime projection did not expose it to clients/AI.

## Implementation
- Runtime unit definitions now include `power_points` from retail `PowerPoints`.
- This preserves scoring/power-rating source data for future execution without claiming complete power-ranking parity yet.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/rules -count=1`
