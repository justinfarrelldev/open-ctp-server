# Task 017 source check: retail AI social sliders

Retail CTP2 strategies include ordered social-slider targets:

- `ctp2/ctp2_released_game/ctp2_data/default/aidata/strategies.txt` has `SliderElement { Delta 3 Food }`, `SliderElement { Delta 3 Production }`, and `SliderElement { Delta 3 Gold }` in `STRATEGY_DEFAULT`.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/strategy.cdb` defines `struct SliderElement { bit(Int) Delta bit(Production) bit(Gold) bit(Food) }` and a `SliderElement[]` array on strategy records.
- `ctp2/ctp2_source_code/ctp2_code/ai/CityManagement/governor.cpp::ComputeBestSliders` reads those elements from the current strategy and assigns production/gold/food deltas.
- `Governor::SetSliders` applies those deltas with `SetWorkdayLevel(expectation - deltaProduction)`, `SetWagesLevel(expectation - deltaGold)`, and `SetRationsLevel(expectation - deltaFood)`, after normalization to the retail slider bounds.

Implementation notes:

- Retail AI strategy parsing now preserves ordered `SliderElement` entries.
- Runtime rules project those elements into `RuntimeAIStrategy`.
- Visibility derives a legal `set-social-policy` action from the first supported strategy slider element and the active government's workday/wages/rations expectations.
- The AI planner consumes non-no-op projected social-policy actions, staying within the projection boundary.
