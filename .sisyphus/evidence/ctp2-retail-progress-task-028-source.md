# Task 028 source check: advance-triggered feats

Retail CTP2 source/data points:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp` checks `advRec->GetTriggerFeatIndex(feat)` when an advance is granted and calls `g_featTracker->AddFeat(feat, m_owner)`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/FeatTracker.cpp` stores active feats by player, tracks duration, and exposes additive effects.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp` applies `FEAT_EFFECT_INCREASE_PRODUCTION` as a production percentage.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp` applies `FEAT_EFFECT_INCREASE_SCIENCE` as a science percentage.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Happy.cpp` applies `FEAT_EFFECT_INCREASE_HAPPINESS` as a flat happiness reason.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp` applies `FEAT_EFFECT_REDUCE_CITY_WALLS` as an additive modifier to city-wall defense for the attacking owner.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/feat.txt` defines the advance-triggered feats, including production/science/happiness bonuses and the gunpowder city-wall reduction.

Server alignment:

- Retail `feat.txt` is imported into runtime feat definitions.
- Retail advances preserve `TriggerFeat` as `trigger_feat_id`.
- Research completion applies the referenced feat as timed nation effects.
- Timed nation effects count down once per round.
- The gunpowder-style `reduce_city_walls` nation effect reduces city-wall defense for the attacking owner without exposing hidden state.
