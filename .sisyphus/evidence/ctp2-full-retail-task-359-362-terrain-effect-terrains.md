# CTP2 Full Retail Tasks 359-362: terrain-improvement terrain restrictions

## Source notes
- Retail `tileimp.txt` `TerrainEffect` blocks list the terrain IDs that each improvement effect applies to.
- Example: `TILEIMP_FARMS` applies to `TERRAIN_DESERT`, `TERRAIN_GRASSLAND`, and `TERRAIN_PLAINS` and requires `ADVANCE_AGRICULTURE`.

## Implementation
- Runtime terrain improvements now preserve `allowed_terrains` from retail `TerrainEffect.Terrain` lists.
- Canonical build-improvement execution rejects improvements on terrain outside the allowed list.
- Safe terrain legal-action projection suppresses disallowed-terrain improvements before AI planning consumes them.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestRetailTerrainImprovementRequiresAllowedTerrain -count=1`
- `go test ./internal/gameplay/visibility -run TestTerrainActionsSuppressDisallowedTerrainImprovements -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
- `go test ./internal/gameplay/ai -count=1`
