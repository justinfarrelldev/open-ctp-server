# Task 048: retail FreeSlaves removes illegal queued slavery items

## Retail source checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:10116-10127`
  - `Player::Emancipate` frees slaves in each city and calls each city's `GetBuildQueue()->RemoveIllegalItems()`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:4183-4187`
  - After `FreeSlaves` exists, units with `SlaveRaids`, `SettlerSlaveRaids`, or `SlaveUprising` are illegal to build.

## Implementation notes

- Emancipation now removes queued production items whose effects are blocked by the global FreeSlaves gate.
- This covers slave raid, settler slave raid, and slave-uprising unit queue entries while preserving legal queued items.
- Existing freeing of city slave specialists and removal of existing slaver/uprising units is unchanged.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-048-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/engine -run 'TestRetailFreeSlavesRemovesQueued|TestRetailFreeSlavesWonderEmancipates' -count=1`
