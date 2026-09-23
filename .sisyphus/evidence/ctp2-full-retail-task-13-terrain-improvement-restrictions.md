# CTP2 full-retail parity task 13 terrain-improvement restrictions evidence

Scope: carry retail terrain-improvement build restrictions into runtime and enforce them when queuing public-works improvements.

Original source/data checked:
- `ctp2_code/gs/newdb/TerrainImprovementRecord.h` includes class, `CantBuildOn`, `Excludes`, construction cost/time, and terrain-effect fields.
- Retail `tileimp.txt` example `TILEIMP_FARMS` has `Class:Farm`, multiple `CantBuildOn` terrains including `TERRAIN_WATER_SHELF`, and class excludes including `Mine`.

Implemented server-compatible slice:
- Runtime terrain improvements now include `Class`, `CantBuildOn`, and `Excludes`.
- Retail projection preserves those fields from typed retail terrain improvements.
- Synthetic/runtime compile can also accept `Class`, `CantBuildOn`, and `Excludes` tags.
- `build-improvement` rejects improvements on forbidden terrain, duplicate completed improvements, and class/ID exclusions against completed improvements on the same tile.

Validation commands:
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TerrainImprovementRuntimeCantBuildAndExcludes -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
