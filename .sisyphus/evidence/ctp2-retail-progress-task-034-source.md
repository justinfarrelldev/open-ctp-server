# Task 034 source check: city founder and recapture feat

Retail CTP2 source/data points:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp` constructor stores the founding owner in `m_founder` when a city is created.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp` conquest/ownership-change code checks `if(m_owner == m_founder)` after ownership changes, adds a city-recaptured score event, and calls `g_featTracker->AddFeat("FEAT_CITY_RECAPTURED", m_owner)`.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/feat.txt` defines `FEAT_CITY_RECAPTURED` with `EffectIncreaseHappiness 1` and `Duration 10`.

Server alignment:

- Canonical `City` state now preserves `FounderID`, initialized to the founding actor in `foundCity` and left stable through later ownership changes.
- `captureCityAt` detects founder recapture when the new owner matches `FounderID` and applies the imported `FEAT_CITY_RECAPTURED` timed nation effect.
- Non-founder captures do not trigger the recapture feat.
- Fullgame observed state hashes were updated because founded city canonical state now records founder identity; event hashes were unchanged.
