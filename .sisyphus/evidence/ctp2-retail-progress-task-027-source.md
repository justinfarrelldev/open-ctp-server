# Task 027 source check: advance-enabled capitalization and infrastructure

Retail CTP2 source/data points:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Advances.cpp` sets player capability flags when an acquired advance has `Capitalization`, `Infrastructure`, or `Transform`:
  - `rec->GetCapitalization()` sets `m_can_build_capitalization = TRUE`.
  - `rec->GetInfrastructure()` sets `m_can_build_infrastructure = TRUE`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp` routes the active build queue head through the special production modes:
  - capitalization computes `ceil(m_shieldstore * government.CapitalizationCoefficient)`, then clears `m_shieldstore`.
  - infrastructure computes `ceil(m_shieldstore * government.InfrastructureCoefficient)`, adds it to the player material pool, then clears `m_shieldstore`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp` adds each city's capitalization gold to the player treasury at turn processing.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Advance.txt` has exactly one released-game `Capitalization` advance (`ADVANCE_GLOBAL_ECONOMICS`) and one `Infrastructure` advance (`ADVANCE_MASS_PRODUCTION`).
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/govern.txt` defines `InfrastructureCoefficient` and `CapitalizationCoefficient` for every government.

Server alignment:

- Retail runtime projection now exposes `CAPITALIZATION` and `INFRASTRUCTURE` production entries gated by their enabling advances.
- Government runtime definitions preserve the two conversion coefficients.
- Server round recalculation keeps these production modes on the queue, converts accumulated city production with `ceil(shields * coefficient)`, credits gold/public works, and clears stored production.
- Safe production projections expose the modes only after the viewer's projected owned advances satisfy the prerequisite.
