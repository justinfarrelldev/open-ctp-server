# CTP2 full-retail parity task 04 AI projected-rules evidence

Scope: move the deterministic server AI one step toward retail-like behavior while preserving the projection-only safety boundary.

Original AI/source checked:
- `ctp2_code/ai/CityManagement/governor.cpp::ComputeBestGovernment`: retail AI filters governments by discovered enabling advances, skips obsolete governments, and then chooses among configured strategy governments.
- `ctp2_code/ai/CityManagement/governor.cpp::ComputeDesiredUnits`: retail AI counts existing/in-progress units, uses strategy build lists, and computes best unit types and desired counts by role/support budget.
- `ctp2_code/ai/CityManagement/governor.cpp` terrain-improvement goal code ranks improvement choices by food/production/gold utility and strategy bonuses.

Implemented server-compatible slice:
- Research selection now uses projected advance costs instead of alphabetical order when costs are visible.
- Government selection now uses projected government unlock advances and projected bonuses instead of only the synthetic hardcoded government.
- Production selection now queues the best visible/unlocked projected unit definition by combat/cargo/settler utility, respecting prerequisites and obsolescence.
- The AI still consumes only `visibility.Projection`; no canonical hidden state or raw retail catalog is exposed to the planner.

Validation command:
- `go test ./internal/gameplay/ai -run 'Projected|RetailLike' -count=1`
