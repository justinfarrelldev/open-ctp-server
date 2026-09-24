# CTP2 Full Retail Task 136: plant-nuke advance prerequisite

## Source evidence
- `advanceutil_Initialize` records the advance with the `PlantNuke` flag.
- `ArmyData::CanPlantNuke` rejects plant-nuke special actions unless the actor has `advanceutil_GetNukeAdvance()`.
- Retail data marks `ADVANCE_NUCLEAR_POWER` with `PlantNuke`.

## Implementation
- Runtime advance definitions now carry `AllowsPlantNuke` / `allows_plant_nuke`.
- Retail runtime projects the `PlantNuke` advance flag.
- Plant-nuke special action rejects actors missing the runtime-declared plant-nuke advance; legacy/synthetic runtimes without such an advance retain prior behavior.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailPlantNukeRequiresNuclearAdvanceWhenRuntimeDeclaresIt -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeIncludesCompleteReleasedCatalog -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`

## Remaining gaps
- Plant-nuke escape/teleport chance and full nuclear blast effects remain future work.
