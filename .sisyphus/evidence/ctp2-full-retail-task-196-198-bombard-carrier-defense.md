# CTP2 Full Retail Tasks 196-198: bombard domains and carrier active-defense enablers

## Source evidence
- `unit.cdb` defines `CanBombard`, `EnableCarrierDefenses`, and `ActiveDefenseOnlyWhenCarryingEnablers`.
- Retail `Units.txt` uses `CanBombard` for artillery/naval/air bombard domains.
- `UnitData::CanActivelyDefend` requires `ActiveDefenseOnlyWhenCarryingEnablers` units to carry at least one cargo unit with `EnableCarrierDefenses` before active defense can fire.

## Implementation
- Retail runtime projection imports `CanBombard` domains as safe unit effects (`can_bombard_land`, `can_bombard_sea`, etc.).
- Bombard targeting now uses `CanBombard` metadata when present, with legacy fallback to prior targetability for old states/tests.
- Retail carrier-defense marker effects import from `ActiveDefenseOnlyWhenCarryingEnablers` and `EnableCarrierDefenses`.
- Active-defense interception now requires a carried enabler when the interceptor declares the retail requirement.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetail(BombardUsesCanBombardDomains|ActiveDefenseRequiresCarrierEnabler)' -count=1`
- `go test ./internal/gameplay/rules -run TestRetailRuntime -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
