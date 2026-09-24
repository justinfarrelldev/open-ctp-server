# CTP2 Full Retail Task 146: plant-nuke escape chance

## Implementation
- Retail `PlantNuke.EscapeChance` now imports as `special_plant_nuke_escape_chance`.
- Successful plant-nuke actions resolve the escape chance after city damage/pollution and before death-chance cleanup.
- Successful escape teleports the acting unit/army to the nearest friendly city tile using deterministic distance and city-ID tie-breaking.
- Zero or missing escape chance preserves the previous no-teleport behavior.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailPlantNukeEscapeChanceTeleportsToNearestOwnCity -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeIncludesCompleteReleasedCatalog -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`

## Remaining gaps
- Full CTP2 nuke blast effects beyond city population/pollution remain future parity work.
