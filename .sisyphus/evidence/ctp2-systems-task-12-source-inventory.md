# Task 12 CTP2 source inventory: public projection/API contracts

Source root inspected before implementation and after the projection/API pass: `/home/justi/Projects/open-ctp-server/ctp2/ctp2_source_code`.

## CTP2 source references
- `ctp2_code/gs/outcom/C3UnitDB.cpp:105` (`CityCanBuildUnit`) exposes unit buildability from public player advances, obsolete advances, movement, and unit records.
- `ctp2_code/gs/outcom/C3UnitDB.cpp:153` / `:160` / `:163` / `:169` expose movement type, production cost, attack, and defense from `UnitRecord` for client/outcom consumers.
- `ctp2_code/gs/newdb/UnitRecord.h:877` / `:905` expose movement-type and enable-advance accessors used by public/gameplay consumers.
- `ctp2_code/gs/gameobj/Advances.cpp:458` / `:468` / `:473` expose `CanResearch` and recompute research availability from owned advances and prerequisites.
- `ctp2_code/gs/gameobj/Player.cpp:4224` gates map visibility through `Player::IsVisible`; public projections must not include unseen canonical map/unit data.
- `ctp2_code/gs/gameobj/Player.cpp:3821` delegates improvement construction through owned cities, and `Player.cpp:5692+` uses player-level diplomatic requests for public diplomacy actions.

## Server alignment
- Server public projection now includes client-safe rules definitions (`rules`), legal command metadata (`legal_actions`), visible unit/city/tile/research/diplomacy state, and stable projected SSE event envelopes.
- Legal action metadata is derived from viewer projection and phase eligibility; wrong-turn/terminal viewers receive no suggested legal actions.
- The projection remains bounded and synthetic; it does not claim full retail CTP2 UI/database parity.
