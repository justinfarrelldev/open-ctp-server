# CTP2 Full Retail Task 69: air-defense bonus

## Source references
- `UnitData::GetDefense` adds `defender.Defense * BonusAirDefense` when the attacker movement type is air.
- `unit.cdb` defines `BonusAirDefense` as a unit behavior flag.

## Implementation
- Runtime unit projection maps `BonusAirDefense` to `combat_defense_vs_air_percent` with the retail boolean value represented as +100% of base defense.
- Combat defense modifiers apply this bonus only against air-domain attackers.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailAirDefenseBonusModifiesDefenderAgainstAirAttackers -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules -count=1`

## Remaining gaps
- Retail active air defense/counter-bombard systems still need separate implementation.
