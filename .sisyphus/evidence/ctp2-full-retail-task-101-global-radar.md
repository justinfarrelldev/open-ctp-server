# CTP2 Full Retail Task 101: global radar visibility

## Source references
- Retail wonder marker `GlobalRadar` grants broad/global map visibility.
- Viewer projections must reveal additional state only to the owning player/nation, not globally.

## Implementation
- Visibility calculation now checks owner/global wonders for `global_radar`.
- Players with global radar mark every tile visible before projection, which safely reveals units/cities through the normal visibility path.
- Other players continue to use normal unit/city sight and fog memory.

## Tests
- `go test ./internal/gameplay/visibility -run TestRetailGlobalRadarRevealsAllTilesAndUnits -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail radar layers, stealth/submarine exceptions, satellite/space interactions, and diplomatic map-sharing remain to be implemented.
