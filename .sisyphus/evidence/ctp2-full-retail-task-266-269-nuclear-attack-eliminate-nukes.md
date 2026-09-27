# CTP2 Full Retail Tasks 266-269: NuclearAttack and eliminate-nukes enforcement

## Retail source
- Retail `UNIT_NUKE` declares `NuclearAttack`.
- Retail production legality checks reject `NuclearAttack` units when nukes are eliminated.

## Implementation
- Retail `NuclearAttack` imports as runtime/canonical `nuclear_attack` marker effect.
- Eliminate-nukes wonder completion removes both `special_plant_nuke` and `nuclear_attack` units.
- Unit production rejects `nuclear_attack` units once the actor's nation has the eliminate-nukes effect.
- Safe production projections suppress `nuclear_attack` units after eliminate-nukes and expose the next legal fallback action.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestRuntimeEliminateNukesBlocksNuclearAttackProduction|TestRetailEliminateNukesWonderDisarmsNationsAndRemovesNukeUnits' -count=1`
- `go test ./internal/gameplay/visibility -run TestProjectionSuppressesNuclearAttackProductionAfterEliminateNukes -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
