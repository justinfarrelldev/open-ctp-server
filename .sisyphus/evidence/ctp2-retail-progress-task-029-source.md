# Task 029 source check: feat field inventory inclusion

Retail data source:
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/feat.txt` contains 14 `FEAT_` records.
- The parser-visible unique field inventory for those records is: `Building`, `Building.Building`, `Description`, `Duration`, `EffectBoatMovement`, `EffectIncreaseCommerce`, `EffectIncreaseHappiness`, `EffectIncreaseHitPoints`, `EffectIncreaseProduction`, `EffectIncreaseScience`, `EffectReduceCityWalls`, `MinimumSizeOfCiv`, and `SlicMessage`.

Server alignment:
- `TestRetailFieldInventorySnapshot` now includes the parsed feat catalog alongside the other released-game gameplay catalogs.
- The field coverage inventory report now accounts for the newly parsed feat catalog and records current runtime coverage status.
