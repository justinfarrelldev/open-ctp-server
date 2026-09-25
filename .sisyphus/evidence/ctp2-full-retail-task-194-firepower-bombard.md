# CTP2 Full Retail Task 194: firepower for bombardment and active defense

## Source evidence
- `UnitData.cpp::FightOneLineDanceRangedAttack` uses `GetFirepower()` when deducting ranged-combat hit points.
- CTP2 active-defense/bombard paths share ranged combat unit stats, so preserving firepower only for melee left ranged damage incomplete.

## Implementation
- Centralized deterministic unit damage calculation with armor floor and firepower multiplier.
- Normal combat, bombardment, and active-defense interception now all use the same firepower-aware damage helper.
- Legacy units with zero firepower continue to behave as firepower `1`.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailFirepowerScales(Combat|Bombard)Damage' -count=1`
- `go test ./internal/gameplay/engine -count=1`
