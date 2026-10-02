# Task 035 source check: feats are one-time achievements

Retail CTP2 source/data points:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/FeatTracker.cpp` stores an `m_achieved` bool array for feat records and serializes it.
- `FeatTracker::AddFeat` returns immediately when `m_achieved[type]` is already true.
- On first successful award, `FeatTracker::AddFeat` sets `m_achieved[type] = true`, adds the active feat, and then links its effects into active effect lists.

Server alignment:

- Canonical `State` now records `AchievedFeatIDs`.
- Runtime feat application records an achieved feat ID on first award and skips all later awards for the same retail feat ID, matching retail global one-time achievement behavior.
- Advance, building, conquered-by-force, and city-recapture feat trigger paths now use this shared one-time gate.
