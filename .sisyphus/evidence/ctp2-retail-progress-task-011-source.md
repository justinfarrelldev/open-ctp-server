# Task 011 source check: government trade route caps

Retail CTP2 exposes `MaxIncomingTrade` and `MaxOutgoingTrade` on governments:

- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/government.cdb` declares both fields as integer government record fields.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/GovernmentRecord.h` exposes `GetMaxIncomingTrade()` and `GetMaxOutgoingTrade()` accessors.
- Released `ctp2/ctp2_released_game/ctp2_data/default/gamedata/govern.txt` assigns both fields for every government; the base-game retail value is `100000`, effectively unlimited for normal games.

Runtime behavior is checked in `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp` around `Player::CreateTradeRoute`:

```cpp
const GovernmentRecord *grec = g_theGovernmentDB->Get(m_government_type);
if(sourceCity.GetOutgoingTrade() >= grec->GetMaxOutgoingTrade()) {
    ...
    return TradeRoute(0);
}

if(destCity.GetIncomingTrade() >= grec->GetMaxIncomingTrade()) {
    ...
    return TradeRoute(0);
}
```

Implementation notes:

- The runtime government projection now carries both caps from `govern.txt`.
- `create-trade-route` rejects new routes when the actor government's outgoing cap for the source city or incoming cap for the destination city is reached.
- Legal action projection suppresses trade-route actions that would be rejected by those caps, preserving the AI/public-projection workflow.
