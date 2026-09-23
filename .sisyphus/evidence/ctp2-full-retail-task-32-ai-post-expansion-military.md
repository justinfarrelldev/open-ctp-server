# CTP2 full-retail parity task 32 AI post-expansion military evidence

Scope: keep retail AI autoplay from consuming all produced settlers without rebuilding any military units.

Implemented:
- Retail-projected AI now prefers combat-capable unit production after expansion when owned non-settler units are fewer than owned cities.
- Combat unit preference runs before city building production, but after active settler expansion production.
- Synthetic/fullgame compatibility remains gated: this post-expansion heuristic is enabled only when projected retail-style `UNIT_SETTLER` rules are present.

Validation command:
- `go test ./internal/gameplay/ai -run 'PlannerQueuesProjectedRetailMilitaryAfterExpansion|PlannerQueuesProjectedRetailSettlerForExpansion|PlannerQueuesProjectedRetailYieldBuildingBeforeUnits' -count=1`
