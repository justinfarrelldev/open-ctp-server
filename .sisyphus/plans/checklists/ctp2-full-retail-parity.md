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
- [x] Task 81: Retail slave-raid special actions execute with slavery protection checks.
- [x] Task 82: Retail create-franchise special actions mark city franchise state.
- [x] Task 83: Retail investigate-city special actions create private city reports.
- [x] Task 84: Retail plant-nuke special actions execute with nuke protection checks.
- [x] Task 85: Retail create-park special actions clean polluted tiles and persist park overlays.
- [x] Task 86: Retail assassinate-ruler special actions destabilize target nations.
- [x] Task 87: Retail force fields block planted nukes.
- [x] Task 88: Retail expel special actions displace adjacent enemy units.
- [x] Task 89: Retail active-defense range stats are preserved in runtime, units, and projections.
- [x] Task 90: Retail active-defense units can intercept bombardment.
- [x] Task 91: Retail injoin special actions block target city production.
- [x] Task 92: Retail reform special actions remove injunction and city crime.
- [x] Task 93: Retail underground-railway special actions free population.
- [x] Task 94: Retail sue special actions apply lawsuit disruption.
- [x] Task 95: Retail food-vat effects reduce city food consumption.
- [x] Task 96: Retail building upkeep is applied as gold maintenance, not production loss.
- [x] Task 97: Retail unit upkeep is preserved and subtracts from national research gold.
- [x] Task 98: Positive city crime reduces production and gold yields.
- [x] Task 99: Retail gold-per-television wonder effects add city gold when a city has television.
- [x] Task 100: Retail boat-movement wonder effects increase sea/amphibious movement refresh.
- [x] Task 101: Retail global-radar wonder effects reveal all tiles through safe projections.
- [x] Task 102: Retail increase-HP wonder effects raise existing and future owner units' health.
- [x] Task 103: Retail embassy wonder effects close protected embassies and project effective global embassies.
- [x] Task 104: Retail free-slaves/prohibit-slavers effects remove slavers and block future slaver use.
- [x] Task 105: Retail polluters-to-parks wonder effects convert the highest-pollution cities to parks.
- [x] Task 106: Retail temporary-full-happiness wonder effects add expiring full-happiness city markers.
- [x] Task 107: Retail reform-cities and empire conversion-prevention wonder effects execute.
- [x] Task 108: Retail spies-everywhere effects can stop technology theft and remove the spy.
- [x] Task 109: Retail eliminate-nukes wonder effects disarm nations and remove nuke-capable units.
- [x] Task 110: Retail protect-from-barbarians wonder effects block barbarian combat.
- [x] Task 111: Retail trade-route wonder gold applies to water and international trade routes.
- [x] Task 112: Retail active-defense bombard interception resolves multiple eligible interceptors deterministically.
- [x] Task 113: Retail trade-route projections are visible only through safe owner/visibility boundaries.
- [x] Task 114: Round recalculation survives city-removing wonder effects without stale slice indexes.
- [x] Task 115: IncreaseBrokerages wonders amplify only active brokerage commerce percentages.
- [x] Task 116: IncreaseScientists wonders add city-scoped research in the current economy model.
- [x] Task 117: Special-action protection checks honor active building/wonder city effects.
- [x] Task 118: Obsolete wonders no longer drive movement refresh or global radar projections.

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
