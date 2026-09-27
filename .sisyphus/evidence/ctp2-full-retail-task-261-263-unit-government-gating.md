# CTP2 Full Retail Tasks 261-263: unit GovernmentType gating

## Retail source
- Retail `Units.txt` uses `GovernmentType` on units such as `UNIT_CLERIC`, `UNIT_TELEVANGELIST`, and `UNIT_FASCIST`.
- `ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp` / `CityData.cpp` check unit government requirements during build legality.

## Implementation
- Runtime unit definitions now preserve retail `GovernmentType` as `required_government_id`.
- Engine unit production rejects government-gated units unless the actor's current government matches.
- Safe production projections suppress wrong-government unit production actions and expose the next legal fallback action.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestRuntimeUnitProductionRequiresCurrentGovernment -count=1`
- `go test ./internal/gameplay/visibility -run TestProjectionSuppressesWrongGovernmentUnitProduction -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
