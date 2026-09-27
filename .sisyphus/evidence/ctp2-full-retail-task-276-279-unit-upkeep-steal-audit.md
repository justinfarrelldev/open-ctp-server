# CTP2 Full Retail Tasks 276-279: unit upkeep/special chance import and field audit

## Implementation
- Preserved retail `FoodHunger` as `food_hunger` on runtime unit definitions, produced canonical units, and owned-unit projections.
- Preserved retail `StealTechnology.SpecificChance` as `special_steal_specific_chance` effect metadata alongside the existing random steal chance and capture/death chance.
- Re-ran the unit field coverage audit against retail `Units.txt`; remaining unprojected unit keys are non-gameplay presentation/sound metadata (`Description`, sprites/icons, sounds, `CargoData.Load`, `CargoData.Unload`) or already represented enum/domain metadata.

## Safety
- Owned-unit projections expose `food_hunger`; visible enemy projections keep upkeep/support/food fields hidden.
- Specific steal chance is preserved as canonical effect metadata for future targeted-tech execution without changing current random steal command semantics.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestProductionCompletionCopiesRuntimeRangedStats -count=1`
- `go test ./internal/gameplay/visibility -run TestTask5ProjectionExposesOwnedUnitCapabilitiesOnly -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
