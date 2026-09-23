# CTP2 full-retail parity task 05 retail runtime projection evidence

Scope: project the source-preserving retail catalog into the existing client-safe runtime-rule shape so clients/AI can see every released-game unit, advance, building, wonder, terrain, and terrain-improvement scalar that the server currently understands.

Original source/schema checked:
- `gs/newdb/unit.cdb`: maps retail movement/cargo/combat/visibility/source flags into safe runtime unit fields while retaining unsupported fields in typed retail raw fields.
- `gs/newdb/advance.cdb`: maps retail costs/prerequisites into runtime advances.
- `gs/newdb/building.cdb` and `wonder.cdb`: map retail city improvements/wonders into production/effect inventories.
- `gs/newdb/terrain.cdb`: maps `EnvBase.Movement` and movement-type bits into runtime movement terrain tables.
- `gs/newdb/TerrainImprovementRecord.h`: maps the first terrain-effect production/yield fields into runtime terrain-improvement definitions.

Validated local released-game projection:
- 74 runtime units, 107 runtime advances, 156 production entries, 26 movement terrains, 40 terrain improvements.
- Selected audited values: `UNIT_SETTLER`, `UNIT_AIRCRAFT_CARRIER`, `ADVANCE_RAILROAD`, `TERRAIN_PLAINS`, and `TILEIMP_FARMS`.

Validation commands:
- `go test ./internal/gameplay/rules -run ProjectRetailRuntime -count=1`
- `go test ./internal/gameplay/rules -count=1`
