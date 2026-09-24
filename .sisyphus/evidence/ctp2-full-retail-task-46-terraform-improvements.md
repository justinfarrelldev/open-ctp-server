# CTP2 Full Retail Task 46: terraform terrain improvements

## Source references
- Retail `tileimp.txt` defines terrain-changing improvements with top-level `TerraformTerrain`, such as `TILEIMP_TERRAFORM_GRASSLAND -> TERRAIN_GRASSLAND`.
- Retail terrain improvements still use `TerrainEffect` blocks for production time/cost and availability metadata.

## Implementation
- Retail terrain improvement catalog import now preserves `TerraformTerrain`.
- Runtime terrain improvements expose safe public `target_terrain` metadata.
- Building a terrain improvement carries `target_terrain` on the queued overlay and changes the base tile terrain when the build completes.
- Completed one-shot terraform overlays are removed after mutating the tile, keeping canonical state compact and deterministic.

## Tests
- `go test ./internal/gameplay/rules -run 'TestLoadRetailRulesFromDirectoryIncludesReleasedCatalog|TestProjectRetailRuntimeMapsSelectedRetailValues' -count=1`
- `go test ./internal/gameplay/engine -run TestRetailTerraformImprovementChangesTileTerrain -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- AI does not yet choose retail terraform actions from `ImprovementLists.txt`; it still only uses the bounded cleanup heuristic.
- Terrain transform material/time fields on base terrain records remain incomplete.
