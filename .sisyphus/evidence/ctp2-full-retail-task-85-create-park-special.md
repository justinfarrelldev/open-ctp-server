# CTP2 Full Retail Task 85: create park special action

## Source references
- Retail unit field `CreateParks` marks park-ranger style units able to create parks/cleanup pollution.
- Retail wonders/buildings include park and polluter-to-park markers that rely on park state.

## Implementation
- Added `create-park` command and park-created event.
- Capable owned units can create a park on their current or adjacent visible tile, spending movement.
- Park creation clears tile pollution and records a completed `park` overlay to prevent duplicates.
- State validation accepts deterministic park overlays.
- Owner legal-action projections expose polluted current/adjacent tiles for capable units.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailCreateParkCleansPollutedAdjacentTile -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail park terrain-improvement yields, ownership rules, AI park planning, and wonder-driven park conversion remain to be implemented.
