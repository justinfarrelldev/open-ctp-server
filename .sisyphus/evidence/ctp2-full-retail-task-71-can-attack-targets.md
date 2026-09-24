# CTP2 Full Retail Task 71: unit CanAttack target restrictions

## Source references
- `unit.cdb` defines typed `CanAttack` bits: land, sea, air, mountain, trade, shallow-water, space, and underwater.
- Retail `Units.txt` declares `CanAttack` target classes per unit.
- CTP2 attack calculation returns zero when a unit cannot attack the defender movement type.

## Implementation
- Retail unit projection maps `CanAttack` classes into runtime effects (`can_attack_land`, `can_attack_sea`, `can_attack_air`, `can_attack_space`, `can_attack_underwater`).
- Retail `IsSubmarine` is projected as a `submarine` effect so underwater-capable attackers can target submarines.
- Combat legality now enforces `CanAttack` metadata when present, while retaining legacy fallback behavior for synthetic/minimal units without target metadata.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailCanAttackEffectsRestrictCombatTargets -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules -count=1`

## Remaining gaps
- Trade-specific attack handling and more exact CTP2 movement-type masks still need deeper implementation.
