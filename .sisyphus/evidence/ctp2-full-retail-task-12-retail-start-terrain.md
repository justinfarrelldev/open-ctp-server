# CTP2 full-retail parity task 12 retail startup terrain evidence

Scope: make retail-runtime startup produce terrain IDs that match the pinned retail movement table.

Original data checked:
- Retail `terrain.txt` IDs include `TERRAIN_PLAINS`, `TERRAIN_FOREST`, `TERRAIN_GRASSLAND`, `TERRAIN_SWAMP`, `TERRAIN_TUNDRA`, `TERRAIN_WATER_DEEP`, and `TERRAIN_WATER_SHELF`.
- The existing world generator emits simplified terrain names (`plains`, `forest`, `grassland`, `swamp`, `tundra`, `ocean`, `shelf`).

Implemented server-compatible slice:
- Synthetic/default startup keeps the previous lowercase terrain mapping.
- Retail-runtime startup maps generated terrain names to retail terrain IDs when the runtime movement table contains `TERRAIN_PLAINS`.
- Engine validation now accepts retail `TERRAIN_*` IDs, and city yield fallback supports the common retail terrain IDs so retail-started cities can recalculate instead of failing on unsupported terrain.

Validation commands:
- `go test ./internal/game -run ServerRuntimePinCanUseRetailProjectionFromEnvironment -count=1`
- `go test ./internal/gameplay/engine -run Terrain -count=1`
- `go test ./internal/game -count=1`
