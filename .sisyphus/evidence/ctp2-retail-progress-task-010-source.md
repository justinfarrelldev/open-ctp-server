# Task 010 source check: science tax capped by government max

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/TaxRate.cpp` implements `TaxRate::SetTaxRates(double s, sint32 owner)` and clamps `s` to `g_theGovernmentDB->Get(g_player[owner]->m_government_type)->GetMaxScienceRate()` before assigning `m_science`.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/govern.txt` defines per-government `MaxScienceRate` values (for example tyranny has `0.6`, monarchy/democracy-style governments vary up to `1`).

Open server now clamps accepted `set-science-rate` commands through the nation government max science effect before persisting the slider value.
