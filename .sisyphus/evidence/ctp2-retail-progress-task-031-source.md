# Task 031 source check: conquered-by-force feat

Retail CTP2 source/data points:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/FeatTracker.cpp` implements `CheckConquerFeat`, looks up `FEAT_CONQUERED_BY_FORCE`, reads `MinimumSizeOfCiv`, and awards the feat to the conquering player when the defeated civilization's maximum city count meets the threshold.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp` calls `g_featTracker->CheckConquerFeat(m_owner, data)` when a player is killed/conquered.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/feat.txt` defines `FEAT_CONQUERED_BY_FORCE` with `Duration 25`, `EffectIncreaseHitPoints 1`, and `MinimumSizeOfCiv 5`.

Server alignment:

- Runtime feat definitions now preserve `minimum_civ_size` from retail `MinimumSizeOfCiv`.
- Nations track `max_city_count` as a canonical, nonnegative state field refreshed during round recalculation.
- Capturing a player's final city applies `FEAT_CONQUERED_BY_FORCE` to the conqueror when the defeated nation's historical max city count meets the retail threshold.
- The existing acceptance state hash was updated because the canonical state now records nation maximum-city-count history.
