# Task 101 source notes — production unlock gates require all prerequisites and any obsolete advance

Retail source evidence:

- `Advances.cpp:501-511` and `Advances.cpp:929-936` iterate every advance prerequisite when determining research availability/needed prerequisites.
- `CityData.cpp:4141-4147` rejects unit production if the unit enable advance is missing or if any listed obsolete advance is known.
- `CityData.cpp:4265-4272` rejects building production if the enable advance is missing or if any listed obsolete advance is known.
- `wonderutil.cpp:300-316` gates wonder availability on its enable advance and `wonderutil_IsObsolete`; `wonderutil.cpp:364-373` treats any listed obsolete advance known by any player as obsoleting the wonder.
- `terrainutil.cpp:394-395` and `terrainutil.cpp:556-557` iterate terrain-improvement obsolete advances, matching the same any-obsolete semantics for public-works actions.

Change:

- `requireProductionUnlock` now checks every runtime prerequisite instead of only the first prerequisite copied into canonical `ProductionItem` metadata.
- Runtime production/unit obsolete lists now reject production if any listed obsolete advance/wonder is active; illegal queue cleanup uses the same helper.
- Added red/green regression coverage for multi-prerequisite building/wonder/unit gates and for second-listed obsolete-advance rejection/cleanup.

Validation:

- RED: `go test ./internal/gameplay/engine -run 'TestRuntimeProductionRequiresAllAdvancePrerequisites|TestRuntimeProductionObsoleteByAnyListedAdvance|TestRuntimeProductionCleanupRemovesItemsAfterSecondObsoleteAdvance' -count=1`
- GREEN: `go test ./internal/gameplay/engine -run 'TestRuntimeProductionRequiresAllAdvancePrerequisites|TestRuntimeProductionObsoleteByAnyListedAdvance|TestRuntimeProductionCleanupRemovesItemsAfterSecondObsoleteAdvance|TestRuntimeBuildingAndWonderProductionUsesPinnedCostAndPrerequisite|TestAdvanceProductionRequiresUnlockedAdvance' -count=1`
- Package: `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`
- Freeze: `go test ./internal/e2e/freeze -run TestEngineLock -count=1`
