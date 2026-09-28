# Task 021 source check: terrain improvement heal rate

Retail source/data evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/terrimprove.cdb` defines `TerrainEffect.HealRate` on terrain improvements.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/TerrainImprovementRecord.h` exposes terrain-improvement effect fields including healing metadata.
- Released terrain-improvement data uses this field for fortification-style improvements that should improve recovery on the tile.

Implementation notes:

- The server already parsed/projected `HealRatePercent` into terrain-improvement overlays.
- Round advancement now heals damaged units standing on completed, owner-compatible improvement overlays by the best tile heal-rate percentage, capped at max health.
- Movement refresh uses the healed health value, matching the existing health-scaled movement model.
