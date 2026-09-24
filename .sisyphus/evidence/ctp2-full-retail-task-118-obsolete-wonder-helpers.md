# CTP2 Full Retail Task 118: obsolete wonder helper behavior

## Implementation
- Boat-movement refresh now uses the same active-wonder applicability rules as other owner wonder effects, including destroyed/obsolete checks.
- Global-radar projection now ignores destroyed/obsolete wonders and respects the wonder owner's obsoleting advances before revealing the map.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailObsoleteBoatMovementWonderDoesNotRefreshSeaUnitMovement -count=1`
- `go test ./internal/gameplay/visibility -run TestRetailObsoleteGlobalRadarWonderDoesNotRevealAllTiles -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- A future cleanup should centralize active-wonder checks across packages instead of duplicating projection-side logic.
