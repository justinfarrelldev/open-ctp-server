# Task 032 source check: feat boat movement effects

Retail CTP2 source/data points:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/FeatTracker.cpp` adds feat records with `EffectBoatMovement` to the `FEAT_EFFECT_BOAT_MOVEMENT` active effect list and exposes them through `GetAdditiveEffect`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp` adds both wonder boat movement and `g_featTracker->GetAdditiveEffect(FEAT_EFFECT_BOAT_MOVEMENT, m_owner)` to sea/shallow-water unit movement at unit initialization and movement recomputation.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/feat.txt` defines `FEAT_SAILED_AROUND_WORLD` with `EffectBoatMovement 100` and `Duration 25`.

Server alignment:

- Timed feat effects are stored as nation effects.
- Server boat movement refresh now sums both active nation `increase_boat_movement` effects and active wonder `increase_boat_movement` effects for sea/amphibious units only.
- Land unit movement remains unaffected by boat movement feat effects.
