# Task 11 CTP2 source inventory: AI legal use

Source root inspected before and after implementation: `/home/justi/Projects/open-ctp-server/ctp2/ctp2_source_code`.

## CTP2 AI model
- `ctp2_code/ai/ctpai.cpp:1367` (`CtpAi::BeginTurn`) changes robot governments through `Governor::ComputeBestGovernment`, applies strategy tax sliders, calls `SetResearch`, refreshes map analysis, and begins diplomacy.
- `ctp2_code/ai/ctpai.cpp:2588` (`CtpAi::SetResearch`) chooses the first strategy-list advance that passes `m_advances->CanResearch`, with active stop-research agreements influencing candidate rejection.
- `ctp2_code/ai/ctpai.cpp:730` / `:805` process AI scheduler and match events; `:932` handles transport-loaded scheduling; `:1138` initializes `SettleMap`, schedulers, and diplomats.
- `ctp2_code/ai/strategy/goals/*` and `ctp2_code/ai/strategy/squads/*` contain the original goal/squad scheduler model.

## Related gameplay APIs used by AI
- `ctp2_code/gs/gameobj/Player.cpp:4164` (`Player::SetResearching`) gates human/network choices through `CanResearch` before changing research.
- `ctp2_code/gs/gameobj/Player.cpp:3821` (`Player::BuildImprovement`) delegates construction only after asserting the city is owned by the player.
- `ctp2_code/gs/gameobj/UnitData.cpp:819` / `:839` gate cargo unload legality before execution by checking movement, army capacity, baddies/ZOC, city capture restrictions, and terrain entry.
- `ctp2_code/gs/gameobj/Player.cpp:5630+` stores and emits diplomatic request/agreement state through public player-level request methods.

## Server alignment
- The server AI remains deliberately bounded and projection-only; it does not emulate retail CTP2 strategy/scheduler competence.
- Task 11 expands deterministic candidate generation to each included server gameplay family using only `visibility.Projection`: research, government, diplomacy proposals/responses, production, terrain public works, transport, combat, founding, embassy, exploration, army creation, fortify, and closeout.
- After implementation, the original source was rechecked at the paths above to verify the server still routes AI decisions through public/legal gameplay commands rather than hidden state mutation.
