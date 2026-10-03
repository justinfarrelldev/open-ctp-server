# Task 106 source evidence: `IsTrader` freight capacity and trade-route transport cost

## Retail source checks

- Released data marks only trader pseudo-units as `IsTrader`:
  - `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Units.txt:500` `UNIT_CARAVAN` has `MaxMovePoints 1`, `MovementType: Trade`, and `IsTrader`.
  - `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Units.txt:4345` `UNIT_Z_FREIGHT_TRANSPORT` has `MaxMovePoints 3` and `IsTrader`.
- Retail creation keeps trader units in a separate list and adds freight capacity instead of placing/visioning them like ordinary map units:
  - `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:1053-1057` inserts `rec->GetIsTrader()` units into `m_traderUnits`, calls `AddTransportPoints((sint32)g_theUnitDB->Get(t)->GetMaxMovePoints())`, and returns before `SetPosition`.
  - `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:215` skips ordinary visibility placement for `GetIsTrader()` units.
  - `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/bldque.cpp:418` exempts traders from full-city-cell build blocking.
- Retail players track total and used freight:
  - `Player.cpp:3271-3278` returns `m_tradeTransportPoints - m_usedTradeTransportPoints` and total `m_tradeTransportPoints`.
  - `Player.cpp:3393-3438` adds/removes total and used transport points.
- Retail route creation and removal use route cost against freight capacity:
  - `Player.cpp:3280-3378` creates a route, calls `PayForTrade`, `AddUsedTransportPoints((sint32)newRoute.GetCost())`, and kills the route if used freight exceeds total freight.
  - `Player.cpp:3440-3450` removes used freight when a route is removed and kills one abstract trader unless the failure cause was no initial caravans.
- Retail path transport cost is path-cost scaled by `CARAVAN_COEF`:
  - `Const.txt:380` sets `CARAVAN_COEF 0.04`.
  - `TradeRouteData.cpp:206-241` accumulates A* path cost, multiplies by `g_theConstDB->GetCaravanCoef()`, rounds, and floors route `m_transportCost` to at least one.

## Server implementation

- Added `CaravanCoefPercent` to retail/runtime constants and import/projection from `CARAVAN_COEF`.
- Added canonical `Nation.TradeTransportPoints` and `TradeRoute.TransportCost` state fields.
- `completeUnitProduction` now handles runtime `is_trader` units by adding movement-derived freight capacity and not appending a normal map `Unit`.
- `createTradeRoute` now calculates route transport cost from visible city endpoints and imported `CARAVAN_COEF`, records it, and rejects creation when the owning nation's public freight capacity would be exceeded.
- Projections expose public used/total freight and suppress over-capacity `create-trade-route` legal actions, so AI route creation remains projection-only.

## Validation

- RED captured in `.sisyphus/evidence/ctp2-retail-progress-task-106-red.txt`.
- Target/package validation:
  - `go test ./internal/gameplay/engine ./internal/gameplay/visibility -run 'TestRetailTraderProductionAddsTransportPointsWithoutMapUnit|TestRetailTradeRoutesConsumeTraderTransportCapacity|TestVisibleCitiesSuppressTradeRouteActionsAtTraderCapacity' -count=1`
  - `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/visibility ./internal/gameplay/ai ./internal/gameplay/http -count=1`
- Freeze/fullgame validation after canonical schema changes:
  - RED: `.sisyphus/evidence/ctp2-retail-progress-task-106-freeze-red.txt`
  - GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-106-freeze-green.txt`
  - Fullgame: `.sisyphus/evidence/ctp2-retail-progress-task-106-fullgame-green.txt`
