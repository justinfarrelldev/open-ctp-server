# CTP2 Full Retail Task 195: active-defense Defend domains

## Source evidence
- `unit.cdb` defines `Defend` domain bits for units.
- `UnitData::CanActivelyDefend` checks the attacker's movement/domain against the defender's `Defend*` bits before allowing active defense.
- Retail data uses this for anti-air/anti-water defenses, e.g. `UNIT_MOBILE_SAM Defend: Air`.

## Implementation
- Retail runtime projection imports `Defend` domains as safe unit effects (`defend_air`, `defend_sea`, etc.).
- Active-defense interception now honors those effects when present, while legacy units without defend metadata retain prior all-domain interception behavior.
- Added validation support for the new defend marker effects.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailActiveDefenseHonorsDefendDomains -count=1`
- `go test ./internal/gameplay/rules -run TestRetailRuntime -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
