# CTP2 Full Retail Task 111: trade-route wonder gold

## Source references
- `Player::CalcWonderGold` in `Player.cpp` adds `GoldPerWaterTradeRoute` for routes crossing water and `GoldPerInternationalTradeRoute` for routes whose destination owner differs from source owner.
- Retail `Wonder.txt` uses these fields for East India Company and Marco Polo's Embassy.

## Implementation
- Added canonical `TradeRoute` state records with stable cloning, sorting, and validation.
- Source city yield now includes owner wonder bonuses for water-crossing and international trade routes.
- Trade-route bonuses are added before building upkeep and crime losses, so existing gold-loss mechanics apply consistently.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailTradeRouteWonderGoldAppliesToSourceCityRoutes -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Route creation/pathfinding, piracy/blockade, goods, capacity, and public safe trade-route projection remain future work.
