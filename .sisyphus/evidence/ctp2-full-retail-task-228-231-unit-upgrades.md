# CTP2 Full Retail Tasks 228-231: unit type identity and upgrade execution

## Implementation
- Canonical units now carry safe `type_id` so runtime-produced and runtime-started units retain their retail unit definition identity.
- Runtime starter definitions now preserve their source unit type, and game startup writes that type onto starter units.
- Runtime unit compilation preserves `UpgradeTo` chains from record definitions.
- Added `upgrade-unit` command and `unit.upgraded` event. Upgrade validation requires ownership, an eligible `UpgradeTo` chain entry, an available target runtime definition, and satisfied target prerequisites.
- Upgrading applies target stats/effects/cargo/support fields, clears fortification, consumes movement for the turn, and preserves proportional health.
- Safe projections expose `upgrade-unit` legal actions only for owned visible units with known type identity and eligible target definitions/prerequisites.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailUpgradeUnit|TestProductionCompletionCopiesRuntimeRangedStats' -count=1`
- `go test ./internal/gameplay/visibility -run TestProjectionExposesRetailUpgradeActionForOwnedUnit -count=1`
- `go test ./internal/gameplay/rules -run 'TestCompileUnitDefinitionPreservesUpgradeChains|TestCompileRuntimePreservesSyntheticUnitCapabilities' -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/ai -count=1`
- `go test ./internal/game -count=1`
