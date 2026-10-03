# Task 046: retail EliminateNukes removes existing nuclear plants and queued plants

## Retail source/data checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:7542-7560`
  - Completing an `EliminateNukes` wonder iterates all players and all cities, calling `CityData::EliminateNukes()`, and removes nuclear-attack units.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:4870-4880`
  - `CityData::EliminateNukes` destroys any built nuclear-plant improvements and calls `m_build_queue.RemoveIllegalItems()`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/buildingutil.cpp:335-339`
  - `buildingutil_IsNuclearPlant` detects buildings flagged `NuclearPlant`.

## Implementation notes

- `eliminateNukes` now removes `EffectNuclearPlant` buildings from every city, mirroring the all-city retail pass.
- Nuclear-plant building queue items are removed at the same time, matching retail's `RemoveIllegalItems()` after the global nuke-elimination wonder.
- Existing nuclear unit/plant-nuke removal and nation nuke-elimination markers remain intact.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-046-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/engine -run 'TestRetailEliminateNukesRemoves|TestRetailEliminateNukesWonderDisarms' -count=1`
