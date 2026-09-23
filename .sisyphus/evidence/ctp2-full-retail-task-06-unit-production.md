# CTP2 full-retail parity task 06 unit production evidence

Scope: make completed city production create a runtime-defined unit at the producing city.

Original source checked:
- `ctp2_code/gs/gameobj/CityData.cpp::FinishBuilding`: city completion delegates to the build queue and resets shield stores when an item is finished.
- `ctp2_code/gs/gameobj/bldque.cpp::BuildQueue::BuildFrontUnit`: unit build completion checks stored production against item cost and queues a `GEV_CreateUnit` event at the city map position.
- `ctp2_code/gs/gameobj/bldque.cpp::BuildQueue::BuildFront`: dispatches unit/building/wonder/endgame completion by front queue category.

Implemented server-compatible slice:
- `RecalculateRound` compiles pinned runtime rules once for the round.
- Completed `ProductionUnit` items now create a deterministic unit at the city tile when the item ID exists in pinned runtime rules.
- Produced units inherit runtime domain, movement, cargo, ZOC, combat, health, visibility, and effects.
- If no runtime unit definition exists, the previous fail-closed no-op behavior is preserved for legacy tests/data.

Validation commands:
- `go test ./internal/gameplay/engine -run ProductionCompletionCreatesRuntimeUnitAtCity -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/e2e/fullgame -run TestPublicExecutorCompletesMinimalScenarioThroughProductionRuntime -count=1`
