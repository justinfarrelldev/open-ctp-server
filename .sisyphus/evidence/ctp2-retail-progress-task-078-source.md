# Task 078 source classification: `GoldHunger` and `FoodHunger`

Retail source/data evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/UnitRecord.cpp:719`, `:1962-1963`, and `UnitRecord.h:1399` parse and expose `FoodHunger`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Unit.cpp:323-326` only returns `FoodHunger` from the unit record; grep found no gameplay call site that consumes food upkeep/consumption from the value.
- `GoldHunger` is absent from the 2003 gameplay source parser/getters; source grep found no `GoldHunger`/`GetGoldHunger` gameplay implementation.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Units.txt:19-22` labels `GoldHunger` as an Apolyton data modification from original Activision code and says it is only activated by user profile / in-game rules screen.
- Released default unit data has all `FoodHunger` values at `0`; `GoldHunger` values are preserved in data but not executed by the original gameplay source.

Classification:

- `FoodHunger` is retail-disabled/no-op in default gameplay: parsed and displayed, but default data is all zero and no gameplay food-consumption path uses it.
- `GoldHunger` is not original retail gameplay in the source release; the server may preserve it as projected metadata/upkeep display, but should not charge treasury upkeep for literal CTP2 retail parity unless a future explicit optional rules toggle is introduced.

Validation evidence:

- Source/data grep and counts: `.sisyphus/evidence/ctp2-retail-progress-task-078-grep.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-078-diff-check.txt`
