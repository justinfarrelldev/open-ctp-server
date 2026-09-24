# CTP2 Full Retail Task 137: watchful city special-action modifier

## Source evidence
- Retail `Const.txt` defines `WATCHFUL_CITY_SUCCESS_MODIFIER 0.5` and `WATCHFUL_CITY_TURNS 7`.
- `CityData::ModifySpecialAttackChance` multiplies special-action chance by the watchful success modifier when `IsWatchful()` is true.
- Retail special-action paths call `SetWatchful()` after attempts such as franchise, plant nuke, slave/underground actions, terror, and conversion-style attacks.

## Implementation
- Added a city `watchful` effect with a seven-turn duration.
- Recalculation decrements and expires watchful city effects.
- Special-action chance helper applies the retail 50% success modifier to already-watchful cities.
- Franchise, plant-nuke, terror, slave raid, underground railway, and conversion attempts now set/reset watchful state after computing the current attempt's chance.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailWatchful|TestRetailSpecialActionAttemptsSetCityWatchful' -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Watchful death-chance modifiers remain future work; retail data sets the death modifier to 1.0, so current retail behavior is unchanged for shipped data.
