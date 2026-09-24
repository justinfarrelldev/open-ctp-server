# CTP2 Full Retail Task 60: retail building placement constraints

## Source references
- Retail `buildings.txt` marks sea-city exclusions with `CantBuildInSea`.
- Retail `buildings.txt` marks coastal improvements with `CoastalBuilding`.

## Implementation
- Runtime production definitions now expose safe placement metadata: `cant_build_in_sea` and `coastal_only`.
- Retail building projection maps `CantBuildInSea` and `CoastalBuilding` flags.
- Server production validation rejects sea-city banned buildings and rejects coastal-only buildings in inland cities.
- Visibility action projection and deterministic AI city production choices apply the same placement filters.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestRuntimeBuildingProductionHonorsSeaAndCoastalPlacement -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/ai ./internal/gameplay/visibility -count=1`

## Remaining gaps
- More specialized retail city/building constraints still need auditing (for example gold-only, capital-only, and any SLIC/scripted constraints).
