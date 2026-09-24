# CTP2 Full Retail Task 61: city walls and wall-ignore combat

## Source references
- `UnitData::GetDefense` removes city-wall defense when the attacker has `IgnoreCityWalls`, then preserves non-wall defender bonuses via `GetDefendersBonusNoWalls`.
- Retail `buildings.txt` marks `IMPROVE_CITY_WALLS` with `CityWalls` and `DefendersPercent`.

## Implementation
- Retail city-wall buildings now project a `city_walls` effect alongside their existing `combat_defense` effect.
- The combat engine subtracts only city-wall defense when an attacker has `ignore_city_walls`, preserving terrain, fortification, and non-wall city defense.
- Runtime unit projection recognizes the `IgnoreCityWalls` flag for modded/variant data even though the stock retail data does not use it.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailIgnoreCityWallsRemovesOnlyWallDefense -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules -count=1`

## Remaining gaps
- Full CTP2 combat still needs force fields, active defense, air-defense bonuses, and the broader stack/order-resolution details.
