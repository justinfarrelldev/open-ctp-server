# CTP2 full-retail parity task 21 pollution turn integration evidence

Scope: wire the retail pollution trigger model into authoritative engine turn advancement.

Original source/data checked:
- `ctp2_code/gs/gameobj/Pollution.cpp::BeginTurn` checks global pollution against the current trigger using strict `>` semantics and advances the trigger phase.
- `ctp2_code/gs/gameobj/Pollution.cpp` executes ozone/flood disasters via `World::OzoneDepletion()` and `World::GlobalWarming(m_gwPhase)`, then increments the global-warming phase for flood disasters.
- `ctp2_code/gs/world/WrldPoll.cpp::GlobalWarming` changes world terrain/water and removes invalid units/cities; `World::OzoneDepletion` marks/pollutes terrain based on UV probabilities.
- Retail `pollution.txt` supplies map-size trigger levels and disaster types.

Implemented server-compatible slice:
- Runtime rules now include projected pollution trigger lists from retail `pollution.txt`.
- Engine state now carries global pollution phase/history/trend and last-disaster metadata.
- `advanceRound` updates pollution using the existing source-modeled `RetailPollutionBeginTurn` and `RetailPollutionEndRound` helpers.
- Until canonical world map size is modeled in engine state, retail trigger selection defaults to `POLLUTION_TRIGGER_LIST_SMALL_MAP` when present.
- Flood disasters increment the stored global-warming phase.

Limitations pending full parity:
- Disaster terrain/city/unit outcomes are not fully executed yet; only trigger/phase/history metadata is wired.
- Pollution inputs are currently bounded to tile pollution totals; city production/population/building/government pollution source aggregation remains pending.
- Map-size-specific trigger selection must switch from the temporary small-map default once world map size is canonical state.

Validation commands:
- `go test ./internal/gameplay/engine -run 'RetailPollutionTriggersDisasterDuringRoundCompletion|RetailPollutionDefaultsToSmallMapTriggerUntilWorldMapSizeIsModeled' -count=1`
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeIncludesCompleteReleasedCatalog -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `GOMAXPROCS=4 go test ./... -count=1`
- `git diff --check`
