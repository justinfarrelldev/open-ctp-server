# Task 023 source check: terrain improvement airports/refuel

Retail source/data evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/terrimprove.cdb` defines `TerrainEffect.Airport` on terrain improvements.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/TerrainImprovementRecord.h` exposes the airport terrain-improvement effect.
- Retail air/fuel mechanics allow air units to refuel at suitable bases; the server already modeled city/transport refuel and projected terrain-improvement airport metadata.

Implementation notes:

- Round advancement now treats completed owner-compatible airport improvement overlays as refuel points.
- Air/fuel units on those overlays refill to `MaxFuel` before no-fuel crash removal is evaluated.
