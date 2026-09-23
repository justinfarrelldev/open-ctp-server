# CTP2 full-retail parity task 31 AI expansion settlers evidence

Scope: improve retail AI autoplay expansion by recognizing projected settler units and founding additional cities after the capital.

Implemented:
- AI now treats free settler units as expansion founders whenever they are not standing on an existing owned city tile.
- AI no longer creates armies from settler units before they can move/found.
- AI production planning prefers an unlocked projected settler unit while it has fewer than three visible owned cities and no active settler, even when beneficial buildings are available.
- AI generated city names now include the unit ID to avoid duplicate city-name rejections after the first city.

Validation commands:
- `go test ./internal/gameplay/ai -run 'PlannerQueuesProjectedRetailSettlerForExpansion|PlannerQueuesProjectedRetailYieldBuildingBeforeUnits|PlannerQueuesProjectedRetailLikeUnitProduction|PlannerSettlesExpansionSettler|PlannerMovesUnescortedSettler' -count=1`
