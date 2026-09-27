# Task 004 source check: hostile combat establishes war and breaks trade

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp` defines `Player::ThisMeansWAR`, which calls `Diplomat::GetDiplomat(m_owner).DeclareWar(defense_owner)` and registers the mutual attack.
- `ctp2/ctp2_source_code/ctp2_code/ai/diplomacy/diplomat.cpp` `Diplomat::DeclareWar` records a declare-war agreement, cancels peace/trade/military/alliance treaties, and calls `g_theTradePool->BreakOffTrade(m_playerId, foreignerId)`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/TradePool.cpp` `TradePool::BreakOffTrade` kills routes between the attack and defense owners.

Open server implementation mirrors this by putting the players into the turn-policy war set, clearing trade/treaty on their relation, and removing direct trade routes when accepted combat begins.
