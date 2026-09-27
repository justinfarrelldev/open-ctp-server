# Task 006 source check: public-works improvements are player/city actions, not worker-only actions

- `ctp2/ctp2_source_code/ctp2_code/gs/outcom/C3InstDB.cpp` exposes `m_player->CanCreateImprovement(...)` and `m_player->CreateImprovement(...)` for terrain improvements/installations, without requiring a worker unit argument.
- `ctp2/ctp2_source_code/ctp2_code/gs/slic/SlicBuiltin.cpp` exposes player public-works tax and public-works level fields (`publicworkstax`, `publicworkslevel`).
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/strategy.cdb` includes AI strategy `PublicWorksPercent`, `PublicWorksReserve`, and terrain improvement lists, showing AI public works are a player-economy loop.

Open server implementation now supports unitless public-works improvement commands in owned/city-radius tiles, projects those legal actions, and lets the AI select them even when no terraforming unit is present.
