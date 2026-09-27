# CTP2 Full Retail Tasks 264-265: upgrade GovernmentType gating

## Implementation
- Runtime upgrade validation now rejects target units whose `required_government_id` does not match the actor's current government.
- Safe upgrade projections suppress wrong-government upgrade actions and expose them once the government matches.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailUpgradeUnitRequiresTargetGovernment -count=1`
- `go test ./internal/gameplay/visibility -run TestProjectionSuppressesWrongGovernmentUpgradeAction -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
