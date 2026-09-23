# CTP2 full-retail parity task 01 source inventory

Scope: add a loss-preserving parser for original CTP2 database text records before projecting them into server-supported runtime rules.

Original source/schema checked before implementation:
- `ctp2_source_code/ctp2_code/gs/newdb/unit.cdb`: unit database grammar includes repeated bit groups, valueless bits, nested special-action structs, cargo, movement, visibility, combat, fuel, and pollution-bearing fields.
- `ctp2_source_code/ctp2_code/gs/newdb/advance.cdb`: advance records include cost, repeated prerequisites, branch/age/icon records, valueless capability bits, pollution modifiers, and feat hooks.
- `ctp2_source_code/ctp2_code/gs/newdb/wonder.cdb`: wonder records include production cost, advance gates, obsolescence, many behavior-bearing valueless/value bits, and pollution effects.
- `ctp2_source_code/ctp2_code/gs/newdb/building.cdb`: building records include production/upkeep, advance gates, and many behavior-bearing bits including pollution modifiers.
- `ctp2_source_code/ctp2_code/gs/newdb/terrain.cdb`: terrain records include nested modifier structs and repeated movement/resource fields.
- `ctp2_source_code/ctp2_code/gs/newdb/government.cdb`: government records include many float/int coefficients, science/rush/trade/readiness values, and advance gates.
- `ctp2_source_code/ctp2_code/gs/newdb/TerrainImprovementRecord.h`: generated terrain-improvement schema confirms class/excludes bit groups and nested effect fields.
- `ctp2_source_code/ctp2_code/gs/database/pollutionrecord.h`, `PollutionRecord.cpp`, `pollutiontoken.h`: pollution text grammar is trigger-list records containing repeated `POLLUTION_TRIGGER level disaster` rows.
- `ctp2_source_code/ctp2_code/gs/gameobj/Pollution.cpp`: runtime global-pollution code accumulates player pollution into history, compares map-size trigger lists, advances disaster phase, and applies flood/ozone events.

Local released-game data used for validation (not copied into the repository):
- `Units.txt`: 74 `UNIT_` records.
- `Advance.txt`: 107 `ADVANCE_` records.
- `Wonder.txt`: 30 `WONDER_` records.
- `gamefile.txt`: retail database load list points the actual city-improvement catalog at `buildings.txt`.
- `buildings.txt`: 52 `IMPROVE_` records.
- `tileimp.txt`: 40 `TILEIMP_` records.
- `terrain.txt`: 26 `TERRAIN_` records.
- `govern.txt`: 12 `GOVERNMENT_` records.
- `pollution.txt`: 4 `POLLUTION_TRIGGER_LIST_` records.

Validation command:
- `go test ./internal/gameplay/rules -run RetailRaw -count=1`
