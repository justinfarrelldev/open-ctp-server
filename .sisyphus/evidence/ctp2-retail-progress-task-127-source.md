# Task 127 source evidence: diplomacy, trade, franchise/legal, capture interactions

## Retail source checks

- Diplomacy/contact/proposals/treaties/regard are handled through `diplomat.cpp`, `PlayerEvent.cpp`, proposal records, embassy effects, and regard drift paths.
- Trade routes use city source/destination goods, agreements, capacities/transport points, piracy, water/international classification, and are cancelled by war/capture (`TradeRouteData.cpp`, `TradePool.cpp`, `tradeutil.cpp`, `PlayerEvent.cpp`).
- Franchise/injunction/lawsuit/expel interactions are implemented as special actions on cities/units with timers and target removal in unit/city event paths (`UnitData.cpp`, `unitevent.cpp`, `CityEvent.cpp`).
- City capture removes/updates ownership-coupled data including trade routes, improvements/wonders/capitol/founder/slave and side effects (`CityEvent.cpp`, `UnitData.cpp`, `wonderutil.cpp`, `buildingutil.cpp`).

## Server coverage

- Diplomacy/contact/proposals/treaties/regard/embassy/close-embassy effects execute or project through public state; legal actions are generated from projections only.
- Trade route creation enforces agreements and caps, assigns goods/gold, consumes trader transport capacity from `IsTrader` units/`CARAVAN_COEF`, classifies water/international routes for wonder gold, supports piracy and treasury payout, and removes routes on war/combat/bombard/capture touching cities.
- Franchise/legal special actions cover franchise production siphon, injunction timers, lawsuit removal, expel/sue interactions, and AI/legal projection without hidden-state shortcuts.
- City capture/recapture paths cover founder identity, conquest distress, trade route removal, improvements/wonders transfer/destruction, slaves, martial law, and regard/war side effects.

## Validation

- GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-127-green.txt`
