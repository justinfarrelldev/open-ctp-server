# Task 033 source check: feat hit-point effects for produced units

Retail CTP2 source/data points:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/FeatTracker.cpp` registers `EffectIncreaseHitPoints` records in the `FEAT_EFFECT_INCREASE_HIT_POINTS` active effect list and exposes them through additive feat effect lookup.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/feat.txt` defines `FEAT_CONQUERED_BY_FORCE` with `EffectIncreaseHitPoints 1` and `Duration 25`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp` already applies retail wonder hit-point bonuses when units are initialized; the server uses the same production-time unit health path for active nation feat hit-point effects so imported feat effects are behavior-bearing rather than parse-only.

Server alignment:

- Active nation `increase_hp` effects now add to produced units' current and maximum hit points along with active wonder hit-point bonuses.
- This makes the conquered-by-force feat's imported `EffectIncreaseHitPoints` runtime effect affect subsequent unit production while preserving deterministic public state.
