# Task 030 source check: building-triggered feats

Retail CTP2 source/data points:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/FeatTracker.cpp` scans feat records with a `Building` block in `FindBuildingFeats` and checks them after `GEV_CreateBuilding` in `FeatBuildingBuilt`.
- `FeatTracker::CheckBuildingFeat` counts the owning player's cities that have the completed building and fires the feat when `Num` or `PercentCities` thresholds are satisfied.
- `FeatTracker::AddFeat` stores active feats and `BeginTurn` removes them after `rec->GetDuration() + feat->GetRound() <= NewTurnCount::GetCurrentRound()`.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/feat.txt` defines building-triggered feats such as `FEAT_THEATERS` (`Building IMPROVE_THEATER Num 8`) and the brokerage/internet/syndicate/orbital-lab variants.

Server alignment:

- Runtime feat definitions now preserve building trigger metadata (`building_id`, `building_count`, and `building_percent`).
- Retail `feat.txt` projection reads same-line building block thresholds used by the released data.
- City production completion checks building feat thresholds after adding the completed building and applies the feat's timed nation effects.
- Building-triggered feats use post-round duration compensation because the server's current `RoundComplete` recalculation adds buildings before the existing timed-effect countdown phase.
