# Task 037 source check: unit SpaceLaunch execution

Retail CTP2 source/data points:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp` `ArmyData::CanLaunch` finds a unit whose DB record has `SpaceLaunch` and can perform a special action.
- `ArmyData::Launch` requires the target map point to contain a valid city owned by the launching army owner; if the launchable unit is in a multi-unit army, retail splits it into its own army before launching.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/unitevent.cpp` `LaunchUnitEvent` reads `SpaceLaunch` turns from the unit DB record, removes the unit from the world, flags it as in-space/left-map, and stores a reentry turn/position.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp` reentry kills the launched units if the target city is no longer valid or no longer owned by the launcher, delays if the destination stack is full, otherwise inserts the units back at the reentry city.

Server alignment:

- Units with `space_launch_turns` effects can execute `space-launch` against an owned destination city tile.
- Launch leaves the map in canonical unit state with `in_space`, `reentry_tile_id`, and `reentry_turns`.
- Round advancement decrements reentry turns, returns the unit to the owned destination city when due, kills it if the city is gone/foreign, and delays if the destination tile is full.
- Launching a unit that belongs to an army detaches it from that army before leaving the map, matching retail's split-before-launch behavior at unit granularity.
- Legal action projection exposes `space-launch` from public owned-city projections for launch-capable units.
