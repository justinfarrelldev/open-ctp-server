# Task 118 source evidence: EnvRiver terrain yield/movement

## Retail source checks

- Retail parses `EnvBase`, `EnvCity`, and `EnvRiver` blocks in `TerrainRecord` (`ctp2_source_code/ctp2_code/gs/newdb/TerrainRecord.cpp:97-98`, `241-243`).
- Retail cell yield calculations add river modifiers when the tile has a river:
  - Food: `Cell.cpp:300-309`
  - Shields/production: `Cell.cpp:338-348`
  - Gold: `Cell.cpp:376-384`
  - Score: `Cell.cpp:433-434`
- Retail movement uses the minimum of base, city, river, and improvement movement costs (`Cell.cpp:858-867`).

## Server implementation

- Added canonical tile `river` state and runtime terrain river fields (`river_food`, `river_production`, `river_gold`, `river_cost`, `river_defense_bonus_percent`) plus city movement cost.
- Retail runtime projection now imports `EnvRiver` modifiers from released terrain data.
- Terrain/city yield paths add river yield modifiers to the selected terrain/city-center yield.
- Terrain movement cost now honors river and city movement costs as minimum costs before tile improvement deltas.
- Combat terrain defense can include river defense modifiers when present.

## Validation

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-118-red.txt`
- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-118-green-target.txt`
- Rules GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-118-rules-green.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-118-green.txt`
- Freeze RED/GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-118-freeze-red.txt`, `.sisyphus/evidence/ctp2-retail-progress-task-118-freeze-green.txt`
- Fullgame GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-118-fullgame-green.txt`
