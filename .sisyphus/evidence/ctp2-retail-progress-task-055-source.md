# Task 055: retail wonder gold is owner treasury income

## Retail source checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:2300-2305`
  - `Player::BeginTurnWonders()` directly adds `CalcWonderGold()` to the player's treasury.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:2324-2368`
  - `CalcWonderGold()` sums `GoldPerWaterTradeRoute`, `GoldPerInternationalTradeRoute`, `GoldPerTelevision`, and `BonusGold` into one owner-level treasury amount.
  - Route wonder gold iterates the trade-pool routes rather than modifying a source city yield.
  - `GoldPerTelevision` counts other players' cities with televisions and multiplies by their population.
  - `BonusGold` is added once per owner, not once per owned city.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Wonder.txt:262-273`
  - Retail data has `IncreaseFoodAllCities 30` for Penicillin and `BonusGold 100` for Pyramids.

## Implementation notes

- Added canonical `bonus_gold` effect kind and mapped retail `BonusGold` to it instead of `yield_gold`.
- `advanceTreasury` now adds owner-level wonder gold directly to national treasury.
- City yield calculation ignores treasury-level wonder-gold effects (`bonus_gold`, route wonder gold, and television wonder gold), preserving normal trade-route city yield separately.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-055-red.txt`
- Green targeted tests:
  - `go test ./internal/gameplay/engine -run 'TestRetail(BonusGold|TradeRouteWonderGold|WonderGoldPerTelevision)' -count=1`
  - `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeIncludesCompleteReleasedCatalog -count=1`
