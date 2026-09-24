# CTP2 Full Retail Task 70: wooden ship attack bonus

## Source references
- `UnitData::CalculateTotalHP`/attack calculation adds `baseattack * AttackWoodenShipBonus` when the defender is a `WoodenShip`.
- Retail `Units.txt` marks wooden vessels and `UNIT_FIRE_TRIREME` with `AttackWoodenShipBonus` / `WoodenShipBonus 0.25`.

## Implementation
- Retail unit projection maps `WoodenShip` and `AttackWoodenShipBonus`/`WoodenShipBonus` into runtime effects.
- Combat attack modifiers now apply the percent bonus against targets carrying the wooden-ship effect.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailWoodenShipAttackBonusAppliesAgainstWoodenShips -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules -count=1`

## Remaining gaps
- Other target-specific attack bonuses (city/submarine/space and similar retail modifiers) still need implementation.
