# CTP2 Full Retail Task 109: eliminate-nukes wonder effect

## Source references
- `Player.cpp` handles `NukesEliminated`/`EliminateNukes` by eliminating nukes across all players' cities and nuclear units.
- Existing Open CTP plant-nuke protection already honors `eliminate_nukes` nation markers.

## Implementation
- Completing an `eliminate_nukes` wonder now upserts an `eliminate_nukes` marker onto every nation.
- Current units with the plant-nuke capability are removed through canonical unit/cargo removal.
- Future planted-nuke specials are blocked by the existing nation-marker check.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailEliminateNukesWonderDisarmsNationsAndRemovesNukeUnits -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Dedicated nuclear missile/warhead records separate from suitcase-nuke special units are not modeled yet.
