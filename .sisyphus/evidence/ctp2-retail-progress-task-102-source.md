# Task 102 source notes — `UpgradeTo` classified and safe optional server command verified

Retail source evidence:

- Released `Units.txt` contains repeated `UpgradeTo` records, but `ctp2_source_code/gs/newdb/UnitRecord.cpp`/`.h` have no `UpgradeTo` parser/accessor token and repository-wide source search finds no unit-upgrade command/cost execution call site.
- Retail executable production legality instead uses `EnableAdvance` plus repeated `ObsoleteAdvance` loops (`CityData::CanBuildUnit`, `Player::CanBuildUnit`) rather than mutating existing units into upgrade targets.
- Therefore `UpgradeTo` is a released-data/mod marker rather than a literal CTP2 retail server gameplay mechanic with source-defined cost semantics.

Server parity stance:

- The server preserves `UpgradeTo` as an optional safe command/projection for compatibility with released data and clients, but no source-backed retail upgrade cost/fuel/HP carryover rule exists to import.
- Existing execution/projection/AI tests verify that optional upgrades fail closed unless the target appears in the chain, target prerequisites/government are satisfied, and the public legal action is projected.
- The runtime upgrade mutation is intentionally deterministic: target definition replaces stats/effects, movement is spent, fuel resets from the target, and HP is preserved proportionally.

Validation:

- `go test ./internal/gameplay/engine -run 'TestRetailUpgradeUnitAppliesTargetDefinitionFromUpgradeChain|TestRetailUpgradeUnitRequiresTargetGovernment|TestRetailUpgradeUnitRejectsUnknownChain' -count=1`
- `go test ./internal/gameplay/visibility -run 'TestProjectionExposesRetailUpgradeActionForOwnedUnit|TestProjectionSuppressesWrongGovernmentUpgradeAction' -count=1`
- `go test ./internal/gameplay/ai -run TestPlannerUsesProjectedUpgradeAction -count=1`
