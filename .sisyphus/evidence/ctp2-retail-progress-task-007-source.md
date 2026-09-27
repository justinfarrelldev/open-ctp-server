# Task 007 source check: retail AI economy science slider

- `ctp2/ctp2_source_code/ctp2_code/ai/ctpai.cpp` reads strategy `PublicWorksPercent` and `SciencePercent`, then calls `player_ptr->SetMaterialsTax(pw_percent/100.0)` and `player_ptr->SetTaxes(sci_percent/100.0)`.
- `ctp2/ctp2_released_game/ctp2_data/default/aidata/strategies.txt` sets `STRATEGY_DEFAULT` to `SciencePercent 60`, `PublicWorksPercent 30`, and `PublicWorksReserve 100`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/TaxRate.cpp` caps science tax by government `MaxScienceRate` when setting tax rates.

Open server now imports strategy economy sliders into runtime rules and projects strategy science-rate legal actions so AI no longer forces the default 100% science/no-gold split when retail strategy asks for 60%.
