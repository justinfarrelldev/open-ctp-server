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
- [x] Task 65: Remaining retail wonder marker fields are projected and preserved.
- [x] Task 66: Retail special-action unit capabilities are projected and preserved.
- [x] Task 67: Retail no-fuel crash units are projected and removed when stranded.
- [x] Task 68: Retail non-capturing units are prevented from capturing cities.
- [x] Task 69: Retail air-defense bonuses are projected and applied in combat.
- [x] Task 70: Retail wooden-ship attack bonuses are projected and applied in combat.
- [x] Task 71: Retail CanAttack target metadata is projected and enforced in combat.
- [x] Task 72: Retail CanEntrench metadata gates fortify actions.
- [x] Task 73: Retail IgnoreZOC metadata bypasses enemy ZOC movement stops.
- [x] Task 74: Retail CanPillage units can pillage completed improvements.
- [x] Task 75: Retail BuildingRemovesAPop units reduce producing city population.
- [x] Task 76: Population-consuming unit production requires spare population.
- [x] Task 77: Retail bombard range/round stats are preserved on produced units and owner projections.
- [x] Task 78: Retail ranged units can execute deterministic bombard actions against visible targets.
- [x] Task 79: Retail convert-city special actions execute with conversion protection checks.
- [x] Task 80: Retail bio/nano/plague special actions execute with protection checks.

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
