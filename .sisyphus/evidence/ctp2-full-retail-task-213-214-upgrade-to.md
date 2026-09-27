# CTP2 Full Retail Tasks 213-214: UpgradeTo field preservation

## Source evidence
- Retail `Units.txt` uses repeated `UpgradeTo` fields (for example `UNIT_ARCHER` upgrades to mounted archers, cavalry, tanks, and fusion tanks).
- `unit.cdb` declares `UpgradeTo` as a repeated unit record field.

## Implementation
- Retail typed catalog now preserves ordered `UpgradeTo` values for units.
- Runtime unit definitions now expose the same ordered upgrade chain for clients/AI/restart-safe runtime projection.
- No unit-upgrade command is claimed complete yet; this closes the data/runtime preservation gap while leaving execution work explicit.

## Tests
- `go test ./internal/gameplay/rules -run 'TestLoadRetailRulesFromDirectoryPreservesTypedRetailValues|TestProjectRetailRuntimeMapsSelectedRetailValues' -count=1`
- `go test ./internal/gameplay/rules -count=1`
