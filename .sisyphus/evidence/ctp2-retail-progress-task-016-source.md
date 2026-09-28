# Task 016 source check: AI public works reserve

Retail CTP2 uses strategy-defined public-works reserve when spending materials on tile improvements:

- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/strategy.cdb` defines `PublicWorksPercent` and `PublicWorksReserve` on strategy records.
- `ctp2/ctp2_source_code/ctp2_code/ai/CityManagement/governor.cpp` calls `strategy.GetPublicWorksReserve(reserve_pw)` before tile-improvement selection.
- The same governor computes `avail_pw = player_ptr->GetMaterialsStored() - reserve_pw`, then only emits `GEV_CreateImprovement` while `spent_pw + needed_pw < avail_pw`.

Implementation notes:

- Retail `strategies.txt` values are already projected into `rules.RuntimeAI`.
- The AI planner now skips projected legal `build-improvement` candidates when the single improvement cost would leave public works/materials at or below the strategy reserve.
- This keeps the projection-driven planner from draining the reserve that retail CTP2 deliberately protects.
