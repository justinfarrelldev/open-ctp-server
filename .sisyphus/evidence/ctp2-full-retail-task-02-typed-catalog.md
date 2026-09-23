# CTP2 full-retail parity task 02 typed catalog evidence

Scope: compile the raw retail database parse into typed, source-preserving catalog structs for every released-game unit, advance, wonder, city improvement, terrain improvement, terrain, government, and pollution trigger list used by `default/gamedata/gamefile.txt`.

Original source/schema checked:
- `gs/newdb/unit.cdb`: authoritative unit stat/bit/group names used for `RetailUnit` fields and preserved raw fields.
- `gs/newdb/advance.cdb`: authoritative cost, repeated prerequisite, and flag names used for `RetailAdvance`.
- `gs/newdb/wonder.cdb`: authoritative production, advance gate, obsolescence, flag, and numeric wonder effect names used for `RetailWonder`.
- `gs/newdb/building.cdb`: authoritative production/upkeep/effect names used for `RetailBuilding`.
- `gs/newdb/terrain.cdb`: authoritative nested `EnvBase` modifier and movement bit names used for `RetailTerrain`.
- `gs/newdb/TerrainImprovementRecord.h`: authoritative tile-improvement class, excludes, construction, and terrain-effect names used for `RetailTerrainImprovement`.
- `gs/newdb/government.cdb`: authoritative government coefficient/rank/advance-gate fields used for `RetailGovernment`.
- `gs/database/PollutionRecord.cpp` and `pollutiontoken.h`: authoritative repeated `POLLUTION_TRIGGER level disaster` parser used for `RetailPollutionTriggerList`.

Validated local released-game catalog counts:
- 74 units, 107 advances, 30 wonders, 52 city improvements, 40 terrain improvements, 26 terrains, 12 governments, and 4 pollution trigger lists.

Validation commands:
- `go test ./internal/gameplay/rules -run RetailRules -count=1`
- `go test ./internal/gameplay/rules -count=1`
