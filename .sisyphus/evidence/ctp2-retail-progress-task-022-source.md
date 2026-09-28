# Task 022 source check: terrain improvement vision/radar projection

Retail source/data evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/terrimprove.cdb` defines terrain-improvement visibility fields, including `CanSee`, `VisionRange`, `Radar`, `RadarRange`, and `ListeningPost` under `TerrainEffect`.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/TerrainImprovementRecord.h` exposes these terrain-improvement records/fields.
- The server already projects these fields into completed `TileOverlay` values.

Implementation notes:

- Visibility now uses completed owner-compatible terrain-improvement overlays as sight sources.
- `VisionRange` and `RadarRange` extend visible tiles from the overlay tile using the existing line-of-sight algorithm; `ListeningPost` supplies at least default sight.
- Pending improvements (`Remaining > 0`) do not reveal tiles.
