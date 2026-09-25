# CTP2 Full Retail Task 182: war discontent owned-fort exemption

## Source evidence
- `PlayHap.cpp::CalcPeaceMovement` counts attacking military units out of cities for war discontent.
- The retail code skips armies on a city tile.
- It also skips armies on a fortification when `terrainutil_HasFort(pos)` is true and the cell owner is the player.

## Implementation
- War discontent now excludes owned attacking units on completed `TILEIMP_FORTIFICATIONS` overlays when the tile owner matches the player.
- Existing city-tile exemption and per-city war-discontent distribution are preserved.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailWarDiscontentExemptsOwnedFortifiedTiles -count=1`
- `go test ./internal/gameplay/engine -count=1`
