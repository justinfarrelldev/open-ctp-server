# Task 020 source check: social slider fit guard

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/ai/CityManagement/governor.cpp::ComputeBestSliders` does not blindly apply the first `SliderElement` from the current strategy. It calls `FitSlidersToCities(sliders_setting)` for each candidate and only remembers the setting when it fits.
- `Governor::NormalizeSliders` clamps proposed workday/wages/rations deltas to retail slider bounds before `SetSliders` applies them.
- `Governor::SetSliders` then applies `WorkdayExpectation - deltaProduction`, `WagesExpectation - deltaGold`, and `RationsExpectation - deltaFood`.

Implementation notes:

- The server's projection layer now applies retail strategy social-slider targets only when visible own-city yields can absorb the candidate policy: production must not stall queued work, projected wages must not make city gold negative, and visible food stores/yields must cover projected ration consumption.
- Without enough visible city fit evidence, visibility keeps the current social policy action payload, which the AI planner treats as a no-op and skips.
- This preserves the retail behavior that strategy sliders are city-fitted rather than blindly applied, and prevents long-game AI production from being throttled by unsafe social-policy changes.
