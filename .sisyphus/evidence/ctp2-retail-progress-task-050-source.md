# Task 050: retail EliminateNukes removes queued nuclear-attack units

## Retail source checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:7542-7559`
  - Completing a nuke-eliminating wonder calls `CityData::EliminateNukes()` for every city and kills existing units whose DB record has `NuclearAttack`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:4870-4881`
  - `CityData::EliminateNukes()` destroys nuclear-plant improvements and calls `m_build_queue.RemoveIllegalItems()`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:4177-4181`
  - `NuclearAttack` units become illegal when the nuke-eliminating global wonder exists, so `RemoveIllegalItems()` removes queued nuclear-attack units too.

## Implementation notes

- Nuke elimination queue cleanup now removes both nuclear-plant building items and nuclear-attack unit items.
- Existing nuclear unit removal and existing/queued nuclear-plant removal are preserved.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-050-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/engine -run 'TestRetailEliminateNukesRemovesQueuedNuclearAttack|TestRetailEliminateNukesRemovesExistingAndQueuedNuclearPlants' -count=1`
