# CTP2 Full Retail Tasks 343-347: nuke city damage constants

## Source notes
- Retail `CityData::GetNuked` kills `NukePopulationPercentage` of city population when population is at least 3.
- Tiny nuked cities are inserted into the kill list and the city tile is killed.
- Retail `Const.txt` defines `NUKE_POPULATION_PERCENTAGE`, `NUKE_KILL_TILES`, and `POLLUTION_CAUSED_BY_NUKE`.

## Implementation
- Imported nuke population, kill-tile, and pollution constants into runtime rules.
- Plant-nuke success now kills the retail configured percentage of city population (75% in retail data).
- Plant-nuke success removes tiny cities with population below 3 and clears attached trade routes.
- Plant-nuke pollution uses the runtime retail nuke pollution constant, with the legacy synthetic fallback retained.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetailPlantNuke(DamagesCityAndHonorsProtection|RequiresNuclearAdvanceWhenRuntimeDeclaresIt|DestroysTinyCity)' -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/visibility -count=1`
