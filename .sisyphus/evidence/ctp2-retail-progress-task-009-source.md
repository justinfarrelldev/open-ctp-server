# Task 009 source check: AI rush-buy gold spending

- `ctp2/ctp2_source_code/ctp2_code/ai/ctpai.cpp` calls `CtpAi::SpendGoldToRushBuy(player)` in the AI turn loop.
- `SpendGoldToRushBuy` gathers cities whose `HowMuchLonger()` is more than one turn, sorts them, checks `GetOvertimeCost()`, and sends `GEV_BuyFront` while enough current savings remain after reserve.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp` implements `PayForBuyFront()` by checking player gold, subtracting `GetOvertimeCost()`, and marking the buy-front state; `BuyFront()` sets the city shield store to the front item cost.
- `CityData::GetOvertimeCost()` computes cost from production remaining and percent remaining (with optional no-rush-buy penalty override).

Open server now exposes a rush-buy command/legal action, charges a deterministic overtime cost based on remaining production and percent remaining, sets stored production to the queued item cost, and lets the normal production pipeline complete it on the next round.
