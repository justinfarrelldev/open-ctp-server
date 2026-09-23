# CTP2 full-retail parity task 33 pollution disaster outcome evidence

Source references reviewed:
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Pollution.cpp`: `BeginTurn` dispatches `POLLUTION_DISASTER_TYPE_OZONE` to `World::OzoneDepletion()` and `POLLUTION_DISASTER_TYPE_FLOOD` to `World::GlobalWarming(m_gwPhase)`.
- `ctp2/ctp2_source_code/ctp2_code/gs/world/WrldPoll.cpp`: flood/global-warming paths convert land to shallow water and remove flooded land cities/armies/improvements; ozone converts affected cells to `TERRAIN_DEAD` and clears improvements/goods.

Implemented bounded deterministic outcomes:
- Flood trigger converts the highest-pollution non-water tile to shallow water (`TERRAIN_WATER_SHALLOW` when retail runtime terrain exists, otherwise synthetic `ocean`).
- Flood trigger clears tile pollution/owner, removes land occupants/cities, prunes empty armies, and removes overlays on the flooded tile.
- Ozone trigger converts the highest-pollution non-water tile to `TERRAIN_DEAD`, clears tile pollution, and removes overlays on the affected tile.
- Existing trigger phase/history behavior remains unchanged.

Validation command:
- `go test ./internal/gameplay/engine -run 'RetailPollution(Flood|Ozone|Triggers|Defaults)' -count=1`
