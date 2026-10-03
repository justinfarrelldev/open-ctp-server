# Task 056: retail CanPirate trade-route piracy

## Retail source checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:2622-2662`
  - `CanInterceptTrade()` requires a trade route in the current cell, the unit DB `CanPirate` flag, enough movement for a special action, and a route whose source owner is not the pirate owner.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:2678-2700`
  - `InterceptTrade()` skips own source-owner routes and schedules `GEV_SetPiratingArmy` for eligible routes.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/tradeevent.cpp:23-35`
  - `SetPiratingArmyEvent` writes the pirating army onto the trade route.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:1058-1074`
  - Pirating armies receive begin-turn gold equal to route value times `GetPiracyWasteCoefficient()` while still on a pirated route cell.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Const.txt:371`
  - Retail `PIRACY_WASTE_COEFFICIENT` is `0.8`.

## Implementation notes

- Added `pirate-trade` command/event and HTTP route.
- `TradeRoute` now stores `PiratingArmyID`.
- Can-pirate units can mark a visible touched foreign-source route as pirated, spending one special-action movement point.
- Begin-turn treasury now pays pirating army owners `route.Gold * PiracyWastePercent / 100` while an army unit remains on a represented route endpoint.
- Legal action projection exposes `pirate-trade` only from live visible route endpoints and public unit effects.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-056-red.txt`
- Green targeted tests:
  - `go test ./internal/gameplay/engine -run TestRetailCanPirate -count=1`
  - `go test ./internal/gameplay/visibility -run TestRetailCanPirate -count=1`
  - `go test ./internal/gameplay/http -run TestHandlerDelegatesPirateTradeRoute -count=1`
  - `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeIncludesCompleteReleasedCatalog -count=1`
