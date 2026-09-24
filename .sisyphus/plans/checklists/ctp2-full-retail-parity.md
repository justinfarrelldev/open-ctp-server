# CTP2 Full Retail Parity Checklist

## Planning
- [x] Current stable work committed before new planning.
- [x] Full retail parity plan created.
- [x] Field coverage inventory created for all retail catalogs.
- [x] AI-vs-AI 100-turn harness created and recording artifact generated.

## Recent Completed Increments
- [x] Task 59: Retail building prerequisite improvements are projected and enforced.
- [x] Task 60: Retail sea/coastal building placement constraints are projected and enforced.
- [x] Task 61: Retail city-wall metadata and wall-ignore combat behavior are implemented.
- [x] Task 62: Retail city protection effects are projected and preserved on completed buildings.
- [x] Task 63: Retail overcrowding-level effects are projected and accumulated on cities.
- [x] Task 64: Remaining retail building marker fields are projected and preserved.

## Retail Data / Runtime
- [ ] Unit catalog behavior field coverage complete.
- [ ] Building catalog behavior field coverage complete.
- [ ] Wonder catalog behavior field coverage complete.
- [ ] Terrain catalog behavior field coverage complete.
- [ ] Government catalog behavior field coverage complete.
- [x] Pollution trigger projection and turn-phase/history integration complete.
- [x] AI list files parsed and runtime-projected.

## Engine Mechanics
- [ ] Full advances/research/unlocks/obsolescence executed.
- [ ] Full unit movement/fuel/refuel/domain/cargo/visibility/stealth executed.
- [ ] Full combat/bombard/active-defense/domain attack rules executed.
- [ ] Full special actions executed.
- [ ] Full city growth/economy/production/upkeep executed.
- [ ] Full building effects executed.
- [ ] Full wonder effects executed.
- [ ] Full terrain improvements/terraform/pillage executed.
- [ ] Full governments/readiness/sliders/support/martial law executed.
- [ ] Full diplomacy/treaties/trade routes executed.
- [ ] Full pollution/disaster outcomes executed.
- [ ] Full victory/endgame/Gaia/score cycle executed.

## AI / Projection / API
- [ ] AI uses retail strategy/build lists from public projections.
- [ ] AI covers research, production, improvements, diplomacy, war, pollution, wonders, endgame.
- [ ] Viewer projections cover every legal action safely.
- [ ] HTTP/SSE command/event contracts cover every mechanic safely.
- [ ] Persistence/restart equality tests cover every mechanic.

## Long-Run Verification
- [x] AI-vs-AI run reaches turn >= 100 and records JSONL transcript.
- [ ] Full retail game can reach terminal victory/final-turn condition.
- [ ] Retail field coverage report has no unimplemented gameplay-affecting fields.
- [ ] Full repository tests pass.
- [ ] Final source/evidence review complete.
