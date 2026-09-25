# CTP2 Full Retail Task 173: readiness HP combat modifier

## Source evidence
- `Readiness.cpp` exposes `GetHPModifier()` from `ReadyPeaceHP`, `ReadyAlertHP`, and `ReadyWarHP`.
- `UnitData.cpp` reduces stored HP by damage scaled with `1 / GetHPModifier()`, making low-readiness units take more effective damage.

## Implementation
- Government readiness HP values now become active nation effects.
- Combat damage against a target is scaled by the target owner's readiness HP modifier.
- Missing/invalid modifiers default safely to 1.0.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailReadinessHPModifierChangesCombatDamage -count=1`
- `go test ./internal/gameplay/engine -count=1`
