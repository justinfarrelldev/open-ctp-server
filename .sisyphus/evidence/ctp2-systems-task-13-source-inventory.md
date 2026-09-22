# Task 13 CTP2 source inventory: cross-system gameplay readiness

Source root inspected before and after the Task 13 cross-system pass: `/home/justi/Projects/open-ctp-server/ctp2/ctp2_source_code`.

## CTP2 source references checked
- `ctp2_code/gs/gameobj/Player.cpp:1835` / `:1960` / `:1993` / `:2082` / `:2383` show turn-begin science, city, production, and full player begin-turn sequencing.
- `ctp2_code/gs/gameobj/Player.cpp:2643` / `:2700` / `:2705` / `:2732` show end-turn/diplomatic request/city/unit turn processing.
- `ctp2_code/gs/gameobj/Advances.cpp:458` / `:468` / `:473` expose research eligibility recomputation used as the reference for per-round research availability.
- `ctp2_code/gs/gameobj/Player.cpp:3794` / `:3818` and `ctp2_code/gs/outcom/C3UnitDB.cpp:107` connect buildability and city production to player advances and unit database records.
- `ctp2_code/gs/gameobj/Player.cpp:763` / `:803`-`:849` / `:906` serialize player-owned units, cities, advances, diplomacy, vision, improvements, material pool, and gaia state for durable restart behavior.
- `ctp2_code/ai/ctpai.cpp` was rechecked for the fact that production AI operates through turn callbacks rather than direct hidden-state mutation in the public-server harness.

## Server alignment
- Policy-based turn completion now advances research just like direct round completion, so client/public turn modes can unlock runtime content deterministically.
- Cross-system Task 13 scenarios use public HTTP/SSE projections and the production AI gateway; they do not use Row 09 or 60-row campaign artifacts.
