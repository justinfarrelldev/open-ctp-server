# CTP2 Full Retail Tasks 348-351: nuke blast tiles and global pollution

## Source notes
- Retail `CityData::GetNuked` removes city-radius units and improvements before adding nuke pollution.
- Retail `Pollution::AddNukePollution` kills `NUKE_KILL_TILES` neighboring tiles and increments pollution history by `POLLUTION_CAUSED_BY_NUKE`.
- Retail `KillTileEvent` changes killed land tiles to `TERRAIN_DEAD` and cuts improvements.

## Implementation
- Successful plant-nuke now removes completed tile improvements in the blast radius while leaving in-progress work intact.
- Retail runtimes kill the configured number of neighboring tiles deterministically through the command RNG domain.
- Killed tiles become `TERRAIN_DEAD` and lose completed improvements.
- Retail nuke pollution is recorded in canonical global pollution history in addition to the existing tile pollution marker used by prior scenarios.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailPlantNuke(KillsConfiguredNeighborTilesAndRecordsGlobalPollution|RemovesCompletedImprovementsInBlastRadius)' -count=1`
- `go test ./internal/gameplay/engine -count=1`
