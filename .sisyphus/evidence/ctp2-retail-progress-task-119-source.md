# Task 119 source evidence: terrain resources, goods, and NeedsCityGood

## Retail source checks

- Retail terrain `Resources` are world-generation inputs: terrain records list up to four resource records (`terrain.cdb:46`), `World::GenerateGoods()` places goods during map generation (`wldgen.cpp:770-869`), and `World::SetRandomGood()` picks resources by terrain probability (`WrlEnv.cpp:313-335`).
- Retail good lookup is canonical world state after generation: `World::GetGood()` resolves a cell's good index through the terrain resource list (`WrlEnv.cpp:247-255`).
- Retail trade/economy uses generated/imported goods from the world/trade route state (`TradePool.cpp:191-196`, `tradeutil.cpp:25-26`, `TradeRouteData.cpp:614-615`).
- Retail units can require a city good via `NeedsCityGood`; released runtime projection imports this requirement for production legality.

## Server coverage/classification

- Server gameplay does not run CTP2 world generation in the gameplay engine; terrain `Resources`/good spawning are classified as world-generation input handled by canonical `Tile.GoodID` supplied by map generation/import.
- Founded cities collect nearby canonical tile goods into `City.Goods`.
- Trade routes carry `GoodID`, calculate good gold, and imported trade-route goods satisfy `NeedsCityGood` production requirements.
- Existing tests lock adjacent goods, local/imported goods, trade-route good selection, and route gold behavior.

## Validation

- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-119-target-green.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-119-green.txt`
