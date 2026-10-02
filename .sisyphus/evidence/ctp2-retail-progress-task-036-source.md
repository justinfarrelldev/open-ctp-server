# Task 036 source check: sailed-around-world feat trigger

Retail CTP2 source/data points:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp` constructs a per-unit `m_roundTheWorldMask` with world X width and sets the starting X bit.
- `UnitData::SetPos` sets the unit's current X bit on movement.
- When all X bits are set and the unit is sea/shallow-water movement, `UnitData::SetPos` calls `g_featTracker->AddFeat("FEAT_SAILED_AROUND_WORLD", m_owner)`.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/feat.txt` defines `FEAT_SAILED_AROUND_WORLD` with `EffectBoatMovement 100` and `Duration 25`.

Server alignment:

- Sea/amphibious units now track visited map X columns in canonical `CircumnavigatedColumns` state.
- Movement records the starting and traversed columns for sea/amphibious movers and awards the imported sailed-around-world feat once all map columns have been visited.
- Land units do not record or trigger the sailed-around-world feat.
- Produced sea/amphibious units initialize their circumnavigation column from their production city tile, matching retail's constructor-start-bit behavior without awarding the feat at construction time.
