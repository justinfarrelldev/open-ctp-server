# CTP2 Full Retail Task 56: terrain transform cost/time data

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/terrainutil.cpp:223-291` computes terraform production time/material cost as source-terrain `TransformRemove` plus target-terrain `TransformAdd`.
- Retail `terrain.txt` defines `TransformAdd` / `TransformRemove` `Time` and `Materials` for base terrain records (for example plains add 1/200 and remove 1/50; forest remove 1/200; grassland add 1/400).

## Implementation
- Typed retail terrain catalog now preserves `TransformAdd` and `TransformRemove` time/material values.
- Runtime terrain projection exposes transform time/material values as safe public terrain metadata.
- Terraform terrain improvements now charge and queue according to source remove plus target add values when both terrain records define transform data, falling back to the improvement defaults otherwise.
- Viewer terrain-improvement legal action generation uses the same projected transform cost/turns for affordability filtering.

## Tests
- `go test ./internal/gameplay/rules -run 'TestLoadRetailRulesFromDirectoryPreservesTypedRetailValues|TestProjectRetailRuntimeMapsSelectedRetailValues' -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetailTerraformImprovementUsesTerrainTransformCostAndTurns|TestRetailTerraformImprovementChangesTileTerrain' -count=1`
- `go test ./internal/gameplay/visibility -run TestTerrainActionsUseRetailAIImprovementListOrder -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `OPEN_CTP_RECORD_RETAIL_AI_TURNS="$PWD/.sisyphus/evidence/ctp2-full-retail-ai-100-turns.jsonl" go test ./internal/e2e/fullgame -run RetailAIAutoPlayRecordsAtLeastHundredTurns -count=1`

## Remaining gaps
- Terraform advance prerequisites, transform partial unit flags, and exact retail worker/army interaction rules remain incomplete.
