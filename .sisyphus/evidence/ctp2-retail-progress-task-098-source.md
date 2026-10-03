# Task 098 source notes — advance map-action flags classified/verified

Retail source evidence:

- `AdvanceRecord.cpp:180-192` and `AdvanceRecord.h:109-121` parse/expose `DeepOcean`, `Transform`, `Roads`, `Railroads`, `MagLev`, and `Tunnels` flags.
- `Advances.cpp:235-247` uses `DeepOcean` to set `Player::SetDeepOceanVisible(TRUE)` and refresh/invalidate the tiled map: this is map presentation/visibility rendering, not server movement/path legality.
- `Advances.cpp:256-258` uses `Transform` to set `m_can_use_terra_tab`, enabling the terraforming UI tab. Actual terraform cost/legality comes from terrain transform data and tile-improvement gates.
- No non-parser retail source call sites were found for `GetRoads`, `GetRailroads`, `GetMagLev`, or `GetTunnels`; released `tileimp.txt` instead gates road/rail/maglev/tunnel-style improvements with per-improvement `EnableAdvance` fields.
- `terrainutil.cpp:247-278` computes terraform time/materials from source terrain `TransformRemove` plus target terrain `TransformAdd` data.

Server parity:

- Runtime projection preserves the advance marker flags for public/rules consumers, but action legality for map improvements uses the source-backed tile-improvement `EnableAdvance`/terrain/cant-build/exclusion data.
- Terrain transform actions compute effective turns/materials from source/target terrain transform data and require the projected tile-improvement `RequiredAdvanceID`.
- `DeepOcean` is classified as presentation/visibility-map metadata; `Transform` as terraforming UI-tab metadata; `Roads`, `Railroads`, `MagLev`, and `Tunnels` as unreferenced parsed markers for server gameplay, superseded by tile-improvement gates in released data.

Validation:

- `go test ./internal/gameplay/engine -run 'TestRetailTerrainImprovementRequiresEnabledAdvance|TestRetailTerraformImprovementUsesTerrainTransformCostAndTurns|TestRetailTerraformImprovementChangesTileTerrain|TestRetailTerrainImprovementRequiresAllowedTerrain|TestTerrainImprovementRuntimeCantBuildAndExcludes' -count=1`
- `go test ./internal/gameplay/visibility -run 'TestTerrainActionsSuppressAdvanceGatedImprovements|TestTerrainActionsSuppressDisallowedTerrainImprovements|TestPublicWorksTerrainActionsSuppressPendingImprovement' -count=1`
