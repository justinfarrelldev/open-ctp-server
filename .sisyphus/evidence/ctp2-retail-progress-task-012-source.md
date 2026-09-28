# Task 012 source check: retail rush-buy modifiers

Retail CTP2 stores per-government rush-buy coefficients:

- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/government.cdb` declares `UnitRushModifier`, `BuildingRushModifier`, `WonderRushModifier`, and `EndGameRushModifier` as float fields.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/GovernmentRecord.h` exposes matching getters.
- Released `ctp2/ctp2_released_game/ctp2_data/default/gamedata/govern.txt` supplies values for every government; e.g. `GOVERNMENT_DEMOCRACY` has `UnitRushModifier 5`, `BuildingRushModifier 4`, `WonderRushModifier 7`, and `EndGameRushModifier 7`.

Retail cost calculation:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/bldque.cpp::BuildQueue::GetTypeCoeff` selects the active government modifier by front queue item category (`UNIT`, `IMPROVEMENT`, `WONDER`, `ENDGAME_OBJECT`).
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp::CityData::GetOvertimeCost` computes:

```cpp
double type_coeff = m_build_queue.GetTypeCoeff();
double percent_remaining = m_build_queue.GetPercentRemaining(m_shieldstore);
sint32 production_remaining = m_build_queue.GetProductionRemaining(m_shieldstore);
sint32 c = sint32(type_coeff * (1 + percent_remaining) * double(production_remaining));
```

Implementation notes:

- Runtime government definitions now project all four retail rush-buy coefficients.
- Engine rush-buy spending uses the actor government's coefficient for the queued item kind while preserving the previous coefficient-1 fallback for synthetic/no-government states.
- Legal action projection uses the same coefficient so AI-visible rush-buy actions do not leak when the player has enough gold for the old synthetic cost but not enough for the retail-modified cost.
