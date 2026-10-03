# Task 047: retail FreeSlaves globally blocks slavery unit production

## Retail source/data checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:4183-4187`
  - City unit production rejects records with `SlaveRaids`, `SettlerSlaveRaids`, or `SlaveUprising` when `wonderutil_GetFreeSlaves(g_theWonderTracker->GetBuiltWonders())` is true.
  - The check uses the global wonder tracker built-wonder mask, not only the builder's own wonders.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:10207-10212`
  - Player production validation applies the same global `FreeSlaves` check.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Wonder.txt:80-86`
  - `WONDER_EMANCIPATION` has `FreeSlaves` and `ProhibitSlavers`.

## Implementation notes

- Slavery production validation now uses a dedicated global FreeSlaves/ProhibitSlavers gate for slave raids, settler slave raids, and slave-uprising units.
- Existing `NoSlaves` local incompatibility remains narrower and still only blocks slaver/settler-slaver units, matching retail's separate checks.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-047-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/engine -run 'TestRetailFreeSlavesWonderGloballyBlocks|TestRetailProhibitSlaversBlocksSlaverProduction|TestRuntimeNoSlaves' -count=1`
