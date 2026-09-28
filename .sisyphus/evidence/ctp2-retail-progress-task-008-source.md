# Task 008 source check: public-works/materials tax diverts production

- `ctp2/ctp2_source_code/ctp2_code/ai/ctpai.cpp` reads `PublicWorksPercent` from the active AI strategy and calls `player_ptr->SetMaterialsTax(pw_percent/100.0)` alongside science tax.
- `ctp2/ctp2_released_game/ctp2_data/default/aidata/strategies.txt` sets `STRATEGY_DEFAULT` to `PublicWorksPercent 30` and `PublicWorksReserve 100`.
- The same AI source periodically reapplies `PublicWorksPercent` and `SciencePercent` when strategies change, so these are ongoing economy sliders rather than one-off starting values.

Open server now has a public-works rate command, projects and plans the retail strategy rate, persists it on the nation, and diverts that share of city production into `Nation.PublicWorks` before applying remaining production to queues.
