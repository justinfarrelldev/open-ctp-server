# Task 9 CTP2 source inventory: diplomacy, trade, governments

Source root inspected before and after implementation: `/home/justi/Projects/open-ctp-server/ctp2/ctp2_source_code`.

## Agreements / diplomacy
- `ctp2_code/gs/gameobj/AgreementData.cpp` models agreements with owner/recipient, serialized state, duration, violation checks, and begin/end turn hooks.
- `ctp2_code/gs/gameobj/DiplomaticRequestData.cpp` creates concrete agreement types such as pollution and piracy pacts through player APIs.
- Server alignment: the bounded slice models contact, pending proposals, accept/reject responses, trade/peace/union treaty state, and canonical restart-stable relation records without copying retail agreement classes.

## Governments
- `ctp2_code/gs/newdb/government.cdb` defines government records; CTP2 player/city systems consult the government database for tax/readiness/economic modifiers.
- `ctp2_code/gs/gameobj/Player.cpp` checks government enable advances and schedules government changes; `TaxRate.cpp`, `Readiness.cpp`, and `CityData.cpp` use the active government for economic coefficients.
- Server alignment: synthetic governments import as runtime definitions with optional advance prerequisites and bounded research/public-works modifiers. `set-government` enforces prerequisite advances and projects only the viewer nation's government.

## World Union
- Existing server World Union support used durable reciprocal embassy state. This task keeps that path and adds accepted union treaties as an equivalent durable relationship prerequisite.
