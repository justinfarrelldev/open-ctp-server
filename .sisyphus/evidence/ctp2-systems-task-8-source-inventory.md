# Task 8 CTP2 source inventory: terrain improvements, public works, and pollution

Source root inspected before and after implementation: `/home/justi/Projects/open-ctp-server/ctp2/ctp2_source_code`.

## Terrain improvements
- `ctp2_code/gs/world/WrlImprove.cpp` `World::InsertImprovement` and `World::RemoveImprovement` store terrain improvements separately from base terrain cells.
- `ctp2_code/gs/world/Cell.cpp` `Cell::InsertImprovement` marks `k_BIT_ENV_HAS_IMPROVEMENT` and stores a terrain-improvement object; `RemoveImprovement` clears the marker after the last improvement is removed.
- Server alignment: synthetic terrain improvements are imported separately from terrain records and compile to runtime build turns, public-works cost, movement/yield deltas, and pollution deltas.

## Public works / build lifecycle
- CTP2 stores improvements as separate map objects; placement and removal are not modeled as base-terrain replacement.
- Server alignment: `build-improvement` queues a deterministic overlay paid from nation public works. Round recalculation decrements build time; movement/yield effects apply only after `Remaining == 0`.

## Pollution
- `ctp2_code/gs/gameobj/Pollution.cpp` `Pollution::GetGlobalPollutionLevel` aggregates global history, player pollution levels, and wonder reductions.
- CTP2 player/agreement code tracks pollution as durable relationship/world state.
- Server alignment: this bounded slice models visible tile pollution counters and cleanup overlays without claiming global-disaster parity; existing world pollution read endpoints remain a separate scaffold.

## Roads/movement and yields
- CTP2 terrain improvements live alongside terrain and influence traversal/economics through world/cell state.
- Server alignment: completed road/farm style overlays modify movement planning and city terrain yields deterministically while hidden pollution remains redacted from public projections.
