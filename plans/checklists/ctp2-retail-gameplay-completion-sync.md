# CTP2 Retail Gameplay Completion Sync Checklist

## Reconciliation

- [x] Saved dirty project-main worktree state before cleanup.
- [x] Saved dirty gameplay-systems worktree state before cleanup.
- [x] Compared dirty main files against gameplay-systems content and kept the newer gameplay-systems branch as the authoritative baseline for overlapping gameplay files.
- [x] Merged main-only `AGENTS.md` guidance into `gameplay-systems-completion`.
- [x] Fast-forwarded `game-server-world`, `gameplay-systems-completion`, and local `main` to the same reconciled commit.
- [x] Verified project root and gameplay worktree were clean after sync.
- [x] Created the new top-level plan and checklist.

## Completed Increments

- [x] Task 001: Long-run retail AI progress test now guards against zero-unit/inert autoplay; AI now queues emergency defenders for populated undefended cities while preserving expansion-settler preference, stale projected combat is suppressed, and combat can target currently visible units without requiring prior fog memory.
- [x] Task 002: Retail AI progress records now include command throughput, active queues, research targets/progress, treasury/public works, trade/diplomacy/war, pollution, combat, terminal, and state-hash metrics.
- [x] Task 003: Retail AI progress records now checkpoint canonical JSON restart equality on every recorded turn.
- [x] Task 004: Accepted combat now establishes mutual war state, clears trade treaty state, and cancels combatant trade routes, matching CTP2 hostile-action diplomacy/trade cleanup.
- [x] Task 005: Trade routes now carry source goods, add source gold yield, and let destination production satisfy NeedsCityGood prerequisites from imported goods.
- [x] Task 006: Public-works terrain improvements can be queued without worker units on owned/city-radius tiles, are projected as legal actions, and are selected by AI planners.
- [x] Task 007: Retail AI strategy economy sliders are imported and the planner uses the strategy science percent instead of forcing 100% science/no-gold economy.
- [x] Task 008: Public-works rate is a gameplay command, AI uses retail strategy public-works percent, and city production is diverted into nation public works before queue progress.
- [x] Task 009: Rush-buy is a legal projected economy action; AI can spend accumulated gold to buy queued production front items into completion via the normal turn pipeline.
- [x] Task 010: Science-rate commands now persist the government-capped science percentage, matching retail tax-rate clamping.
- [x] Task 011: Government trade caps from retail `govern.txt` are projected into runtime rules, enforced when creating trade routes, and suppress invalid projected legal actions.
- [x] Task 012: Retail government rush-buy coefficients now project from `govern.txt`, engine rush-buy costs use the active government/item type modifier, and legal actions suppress unaffordable retail-modified rush buys.
- [x] Task 013: AI planners now consume projected legal trade-pact proposal actions so trade agreements can precede goods trade routes without hidden-state shortcuts.
- [x] Task 014: AI planners now consume projected legal peace proposal actions so combat-created wars have a public-projection treaty recovery path.
- [x] Task 015: AI planners now consume projected legal bombard actions so bombard-capable units participate in combat planning without hidden-state shortcuts.
- [x] Task 016: AI terrain-improvement planning now preserves retail `PublicWorksReserve` instead of spending public works/materials down to zero.
- [x] Task 017: Retail AI `SliderElement` social-policy deltas are imported, projected, exposed as legal actions, and consumed by AI planning.
- [x] Task 018: AI planners now consume projected legal `upgrade-unit` actions so retail `UpgradeTo` chains can advance autonomous units.
- [x] Task 019: AI planners now consume projected retail special city/tile actions for soothsaying and park creation.
- [x] Task 020: Retail AI social-slider projection now requires visible city fit evidence before changing workday/wages/rations, matching governor fit checks and preserving long-game production.
- [x] Task 021: Terrain improvement `HealRate` now executes during round advancement by healing owner-compatible units on completed improvement overlays.
- [x] Task 022: Terrain improvement `VisionRange`/`RadarRange`/`ListeningPost` now extend player visibility from completed improvement overlays.
- [x] Task 023: Terrain improvement `Airport` now acts as an owner-compatible refuel point during round advancement.
- [x] Task 024: AI now suppresses fragile legacy `terraform` unit commands and relies on validated public-works `build-improvement` terraform paths instead.
- [x] Task 025: Trade-route legal actions now require current visible source/destination cities, preventing stale fog-memory routes from reaching AI submission.
- [x] Task 026: Public conquest capture recovery now keeps reserve capture units out of combat, switches to visible mobile legal reserves when needed, routes remembered city targets one projected legal step at a time, and avoids stale-memory frontier oscillation.
- [x] Task 027: Advance-enabled capitalization and infrastructure production now projects from retail advances/government coefficients, remains public-advance gated, and converts city production into gold or public works with retail ceiling semantics.
- [x] Task 028: Retail `TriggerFeat` advances now import feat definitions, apply timed nation effects on research completion, and execute the gunpowder city-wall reduction path.
- [x] Task 029: Retail feat catalog fields are included in the field inventory snapshot and coverage report now that `feat.txt` is parsed for advance-triggered feats.
- [x] Task 030: Retail building-triggered feats now import `Building` thresholds and apply timed nation effects when city production completes the required building count.
- [x] Task 031: Retail `FEAT_CONQUERED_BY_FORCE` now imports `MinimumSizeOfCiv`, tracks nation max city count, and applies the hit-point feat when a qualifying civilization loses its final city.
- [x] Task 032: Retail feat boat-movement effects now contribute to sea/amphibious movement refresh alongside wonder boat-movement bonuses.
- [x] Task 033: Retail feat hit-point effects now add to produced units' current and maximum hit points while active.
- [x] Task 034: Retail city founder identity is preserved and `FEAT_CITY_RECAPTURED` is awarded when a founder retakes a city.
- [x] Task 035: Retail feats now record global achievement IDs and cannot be awarded repeatedly after first success.
- [x] Task 036: Retail `FEAT_SAILED_AROUND_WORLD` now triggers from sea/amphibious unit circumnavigation columns.
- [x] Task 037: Retail unit `SpaceLaunch` now launches units to owned cities and resolves delayed reentry/failed reentry.
- [x] Task 038: Space-launch commands are exposed through the authenticated HTTP gameplay route with safe payload fields.
- [x] Task 039: Retail `NuclearAttack` units can directly nuke vulnerable foreign cities and are consumed on use or protection.
- [x] Task 040: Retail `ThrowParty` units now spend gold/movement, respect target reception cooldown, and apply timed diplomacy regard through projection and HTTP.
- [x] Task 041: Retail wonder `IncreaseRegard` now improves public foreign regard toward active wonder owners.
- [x] Task 042: Retail `Advertise` units now execute the dedicated advertise route by reusing CauseUnhappiness amount/timer/chance mechanics.
- [x] Task 043: Retail `EnableParkRangers` now globally unlocks `CreateParks`/park-ranger unit production as in the built-wonders mask check.
- [x] Task 044: Retail `DoubleTelevangelists` now gates enhanced televangelist conversion gold; unboosted televangelists fall back to cleric conversion.
- [x] Task 045: Retail `EliminateNukes` now globally blocks `NuclearPlant` building production.
- [x] Task 046: Retail `EliminateNukes` now destroys existing `NuclearPlant` buildings and removes queued nuclear-plant items in all cities.
- [x] Task 047: Retail `FreeSlaves` now globally blocks slave raid, settler slave raid, and slave-uprising unit production.
- [x] Task 048: Retail `FreeSlaves` emancipation now removes queued slave raid, settler slave raid, and slave-uprising unit items.
- [x] Task 049: Retail `EliminateNukes` now globally blocks nuclear-attack unit production even in imported states with pre-existing defuser wonders.
- [x] Task 050: Retail `EliminateNukes` now removes queued nuclear-attack unit items when the defuser effect runs.
- [x] Task 051: Retail `InvisibleForGlobeSat` units are hidden from global-radar-only visibility while remaining visible to normal sight.
- [x] Task 052: Retail disembark now zeroes movement for non-air/non-beach-assault cargo while preserving beach-assault unload movement.
- [x] Task 053: Retail `CanBeGifted` units now execute/projection/HTTP gift-unit actions that transfer ownership at the giver capitol.
- [x] Task 054: Retail `DeathEffectsHappy` now gates combat peace-movement lost-unit counters instead of all combat casualties counting.
- [x] Task 055: Retail wonder gold (`BonusGold`, route gold, and television gold) now pays owner treasury once at begin turn instead of inflating city yields.
- [x] Task 056: Retail `CanPirate` units now mark trade routes as pirated and earn route-value piracy gold at begin turn.
- [x] Task 057: Retail Gaia countdown now requires configured computers, satellites, processing towers, and tower coverage.
- [x] Task 058: Retail `VETERAN_COEF` now increases veteran melee, bombard, active-defense, and counter-bombard hit strength.
- [x] Task 059: Retail wonder `BuildingEverywhere`/`ActualBuildingEverywhere` now provide effective buildings for owner city building prerequisites.
- [x] Task 060: Retail `GoldPerTelevision` wonder income now counts television effects from effective buildings, including building-everywhere wonders.
- [x] Task 061: Retail city capture now transfers built wonders to the conqueror and reapplies owner-facing wonder add effects.
- [x] Task 062: Retail city capture now destroys captured Capitol buildings while preserving other improvements.
- [x] Task 063: Retail city capture now clears owner-specific city effects and timed happiness/special effects.
- [x] Task 064: Retail bombard can target undefended cities, importing city-bombard constants and applying building/population damage with stale-safe projections.
- [x] Task 065: Retail city bombard building destruction now removes queued production that became illegal.
- [x] Task 066: Retail city bombard wall-defense probability now uses actual city wall buildings only.
- [x] Task 067: Retail `IsFlanker` now affects melee row targeting so flankers can attack nearby occupied rows when same-row targets are gone.
- [x] Task 068: Retail ranged combat now prefers same/nearest tactical-row targets before fallback.
- [x] Task 069: Retail bombard hostilities now start war and cancel trade routes between combatants.
- [x] Task 070: Retail `LossMoveToDmgTwo` now imports and refreshes damaged movement with a minimum of two points.
- [x] Task 071: Retail counter-bombard now considers every living defender on the target tile, not only the attacker's eligible bombard targets.
- [x] Task 072: Retail counter-bombard now honors the defender's own bombard range to the original bombarder.
- [x] Task 073: Retail active-defense range is now measured from interceptor to bombarder, not interceptor to target tile.
- [x] Task 074: Retail `Explodes` is source-classified as presentation-only combat explosion animation metadata.
- [x] Task 075: Retail `HasPopAndCanBuild` city pseudo-units are rejected from normal city production queues.
- [x] Task 076: Retail active-defense interceptors now fire once per turn and reset on turn refresh.
- [x] Task 077: Retail active-defense interceptors now resolve in descending attack order.
- [x] Task 078: Retail `GoldHunger` and `FoodHunger` are source-classified as non-executed/default-disabled upkeep metadata.
- [x] Task 079: Government `ProfessionalUnits` is source-classified as zero-valued default AI utility metadata; per-unit `IsProfessional` remains the gameplay support exemption.
- [x] Task 080: Retail readiness support now floors after summing fractional per-unit costs, including disband budget checks.
- [x] Task 081: Building `AllowGrunts` is source-classified as non-executed/AI-placeholder metadata.
- [x] Task 082: Unit `NoIndex`, `GLHidden`, `DefaultIcon`, `DefaultSprite`, and `Sound*` fields are source-classified as presentation/setup metadata.
- [x] Task 083: Retail cargo `CanCarry` now imports exact size/domain classes and filters engine/legal/AI embark by class compatibility.
- [x] Task 084: Retail cargo `Load`/`Unload` nested sound records are source-classified as presentation-only metadata, and transport death cargo removal is verified.

## Full-Game Mechanical Execution First

- [x] Add/restore a RED long-run retail AI progress test that fails when autoplay becomes inert, loses all units, stops producing/researching, or issues only round-complete commands late-game.
- [x] Extend the full-game progress record with cities, units, active queues, research progress, advances, gold/treasury, public works, trade routes, diplomacy/war, pollution, combat counts, command throughput, terminal state, and state hash.
- [x] Make the progress harness checkpoint canonical restart equality.
- [x] Fix AI economy/production continuity using only public projections and legal actions.
- [x] Fix AI exploration, settling, worker/public-works improvement, transport/cargo/fuel dead ends, and pathing stalls.
- [x] Fix coarse diplomacy/trade/goods loops so trade agreements, routes, goods-gated production, and gold flow materially affect long games.
- [x] Fix coarse combat loops for war/contact, legal targeting, bombardment/active defense, ZOC, stealth detection, city defense/capture, and conquest progress.
- [x] Fix pollution/disaster/government-slider/readiness/happiness/crime/endgame loops for long-game stability.

## Retail Parity Closure

- [ ] Complete runtime execution for all gameplay-affecting advance fields.
- [ ] Complete runtime execution for all gameplay-affecting unit fields and actions.
- [ ] Complete runtime execution for all gameplay-affecting wonder fields.
- [ ] Complete runtime execution for all gameplay-affecting building fields.
- [ ] Complete runtime execution for all terrain, worker, public-works, goods, and trade effects.
- [ ] Complete runtime execution for all government, diplomacy, slider, readiness, happiness, crime, and pollution effects.
- [ ] Complete runtime execution for all victory/endgame/space/Gaia/final-turn mechanics.
- [ ] Update the retail field coverage report so no behavior-bearing retail field remains unimplemented or parse-only.

## Detailed Remaining Retail Parity Ledger

This ledger is the authoritative remaining-task checklist. It must be kept current: when a slice is implemented, mark or refine the corresponding checkbox and add the concrete completed task under "Completed Increments" / numbered task history with source evidence.

### Source coverage and classification

- [ ] Re-audit every released `gamedata` catalog field against `ctp2/ctp2_source_code` call sites and classify each as executed gameplay, safely projected marker, presentation-only, AI-only, or unsupported bug-fixed behavior.
- [ ] Update `.sisyphus/evidence/ctp2-full-retail-task-22-field-coverage-inventory.md` so no gameplay-affecting field remains "parse-only" or ambiguously classified.
- [ ] Add/maintain regression tests that fail when newly parsed retail fields are omitted from the coverage/classification ledger.
- [ ] Import and execute remaining gameplay constants from `Const.txt` that affect modeled mechanics, including fuel, trade-route path cost, special-action costs/chances, unrest, and disaster constants.

### Advances and research

- [ ] Verify and close all advance unlock gates for unit/building/wonder/terrain-improvement production, including multi-prerequisite edge cases and obsolete-advance removal.
- [ ] Complete or classify `Age`, `Branch`, `Icon`, `KeyAdvanceAdvice`, and `GLHidden` as presentation/AI metadata where source confirms no gameplay effect.
- [ ] Finish source-backed execution/verification for `DeepOcean`, `Roads`, `Railroads`, `MagLev`, `Tunnels`, and `Transform` map-action gates.
- [ ] Verify `PlantNuke`, `Capitalization`, `Infrastructure`, `PollutionProductionModifier`, `PollutionSizeModifier`, and `TriggerFeat` parity across government/advance transitions and persistence.

### Unit catalog fields, movement, support, cargo, and visibility

- [x] Execute retail material support completely: `ShieldHunger`, `NeedsNoSupport`, `IsSpecialForces`, readiness support coefficients, fractional cost rounding, disband ordering, and wonder readiness reductions.
- [x] Source-classify `ProfessionalUnits` government quota/behavior as default-zero AI utility metadata; per-unit professional support/disband behavior is implemented.
- [x] Source-check and close `GoldHunger` and `FoodHunger`: classified as non-executed/default-disabled upkeep metadata for literal retail parity.
- [ ] Complete `IsTrader` behavior: trader unit lists, trade transport points, route capacity, route path-cost consumption, trader death/removal, and AI route creation.
- [x] Execute retail cargo semantics for `CargoData.MaxCargo`, `CanCarry`, domain/size constraints, engine embark, legal public projection, AI embark planning, transport-destruction cargo fate, and `CargoData.Load`/`CargoData.Unload` presentation-only classification.
- [ ] Finish air/space/sea/land/domain movement parity, including `MovementType`, `MaxMovePoints`, `MaxFuel`, `NoFuelThenCrash`, `LossMoveToDmgNone`, non-space fuel costs, crash/damage rules, city/airport/carrier refuel, and hostile terrain immunity.
- [ ] Verify `CanBeachAssault`, `Paratrooper`, and `ParatrooperTransport` beyond unload movement: legal target selection, transport requirements, and post-drop movement/combat restrictions.
- [ ] Complete `VisionRange`, `VisionClass`, `CanSee`, `IsSubmarine`, `Stealth`, `CanSeeStealth`, `InvisibleForGlobeSat`, global radar, radar/listening-post interactions, and stale-fog legal action suppression.
- [ ] Verify `NoZoc`/`IgnoreZOC`/ZOC movement stops and AI pathing behavior against retail.
- [ ] Complete unit production restrictions: `BuildingRemovesAPop`, `CantBuild`, `OnlyBuildOne`, `GovernmentType`, `NeedsCityGood`, `NoSlaves`, `EnableAdvance`, `ObsoleteAdvance`, and `UpgradeTo`; `HasPopAndCanBuild` city pseudo-units are rejected from normal queues, and `NoIndex`/`GLHidden` are presentation/setup metadata.
- [ ] Complete upgrade semantics for `UpgradeTo`: costs, prerequisites, HP/fuel/movement carryover, legal projection, HTTP, AI use, and persistence.
- [x] Classify all unit sound/icon/sprite fields (`DefaultIcon`, `DefaultSprite`, `Sound*`) as presentation/setup metadata for server gameplay.

### Unit combat and special actions

- [ ] Implement or source-classify `IsFlanker` tactical targeting, including grouped-combat target row/column behavior.
- [x] Source-classify `Explodes` death/combat behavior as presentation-only battle explosion animation metadata.
- [ ] Verify all combat stats and modifiers: `Attack`, `Defense`, `ZBRangeAttack`, `BombRounds`, `BombardRange`, `Firepower`, `Armor`, `MaxHP`, `MountedBonus`, `BonusAgainstMounted`, `AttackWoodenShipBonus`, `WoodenShipBonus`, `Defend`, `CanAttack`, `CanBombard`, `CanCounterBombard`, `CityWalls`, `ReduceCityWalls`, and `IgnoreCityWalls`.
- [ ] Complete active defense parity: `ActiveDefenseRange`, `ActiveDefenseOnlyWhenCarryingEnablers`, `EnableCarrierDefenses`, interceptor target selection, once-per-turn use/reset, enabler cargo checks, and bombard/counter-bombard RNG ordering.
- [ ] Complete `Civilian`, `CantCaptureCity`, `VictoryEnslavement`, `DeathEffectsHappy`, `SingleUse`, and `NuclearAttack` edge cases under grouped combat/city capture.
- [ ] Complete all spy/special-action vectors field-by-field: `EstablishEmbassy`, `StealTechnology`, `PlantNuke`, `InvestigateCity`, `AssasinateRuler`, `BioTerror`, `NanoTerror`, `Plague`, `ConvertCities`, `CreateFranchise`, `CauseUnhappiness`, `Advertise`, `InciteRevolution`, `IndulgenceSales`, `SlaveRaids`, `SettlerSlaveRaids`, `SlaveUprising`, `UndergroundRailway`, `CanInjoin`, `CanReform`, `CanSue`, `CanExpel`, `CreateParks`, `ThrowParty`, and nested `Chance`/`DeathChance`/`EliteChance`/`EscapeChance`/`Effect`/`Timer`/`Amount`/`Sound` fields.
- [ ] Ensure every special action has safe public legal-action projection, HTTP route coverage where applicable, canonical event payloads, AI-consumable legality, and no hidden-state shortcut.

### Buildings and city effects

- [ ] Finish all building yield/economy effects: `FoodPercent`, `ProductionPercent`, `CommercePercent`, `SciencePercent`, `GoldPerCitizen`, `Upkeep`, and interactions with sliders/government coefficients.
- [ ] Finish population/happiness effects: `HappyInc`, `RaiseMaxPopulation`, `RaiseOvercrowdingLevel`, `FoodVat`, `StarvationProtection`, `Cathedral`, `IsReligious`, `LowerPeaceMovement`, and temporary city effects.
- [ ] Finish crime/pollution effects: `LowerCrime`, `PopulationPollutionPercent`, `ProductionPollutionPercent`, `PollutionAmount`/`Pollution`, `NuclearPlant`, and wonder interactions.
- [ ] Finish protection effects: `ProtectFromNukes`, `ProtectFromBioAgents`, `ProtectFromNanoVirus`, `PreventConversion`, `PreventSlavery`, and `ForceField` for both direct buildings and effective buildings.
- [ ] Finish military/building effects: `DefendersPercent`, `OffenseBonusLand`, `OffenseBonusWater`, `OffenseBonusAir`, `CityWalls`, `Airport`, and effective-building consequences; `AllowGrunts` is source-classified as non-executed/AI-placeholder metadata.
- [ ] Finish structural/build-gate effects: `PrerequisiteBuilding`, `CoastalBuilding`, `CantBuildInSea`, `CantBuildOnLand`, `Capitol`, `Brokerage`, `Television`, `DoubleTelevangelists`, `DefaultIcon`, and `Description` classification.
- [ ] Ensure wonder-granted effective buildings participate in every retail `GetEffectiveBuildings()` use: yields, science, happiness, crime, pollution, protection, refuel/airport, city walls/defense/offense, television, brokerage, and production prerequisites.

### Wonders

- [ ] Finish numeric wonder effects: `IncKnowledgePercent`, `IncreaseFoodAllCities`, `IncreaseProduction`, `IncHappinessEmpire`, `DecCrimePercent`, `DecEmpireSize`, `BonusGold`, `GoldPerWaterTradeRoute`, `GoldPerInternationalTradeRoute`, and `GoldPerTelevision` including obsolescence/destruction.
- [ ] Finish global/diplomacy wonder effects: `CloseEmbassies`, `EmbassiesEverywhereEvenAtWar`, `IncreaseRegard`, `SpiesEverywhere`, `ProtectFromBarbarians`, `RevoltingCitiesJoinPlayer`, and relation/war/embassy edge cases.
- [ ] Finish production/prohibition wonder effects: `EnableParkRangers`, `FreeSlaves`, `ProhibitSlavers`, `EliminateNukes`, `ReformCities`, `PollutersToParks`, and illegal queue cleanup/reversal edge cases.
- [ ] Finish happiness/readiness/pollution wonder effects: `TemporaryFullHappiness`, `NoPollutionUnhappiness`, `ReduceReadinessCost`, and source-backed duration/stacking behavior.
- [ ] Finish visibility/endgame wonder effects: `GlobalRadar`, `StartGaiaController`, `BuildingEverywhere`, `ActualBuildingEverywhere`, obsolescence, destruction, capture, and transfer behavior.
- [ ] Classify wonder `DefaultIcon`, `Description`, and `Movie` as presentation-only unless public contract requires projection.

### Terrain, tile improvements, goods, and map rules

- [ ] Finish terrain `EnvBase`, `EnvCity`, and `EnvRiver` yields/movement/score/defense, including river overlays and city-center yield selection.
- [ ] Implement or classify terrain `Resources`, goods spawning, goods trade bonuses, `NeedsCityGood`, and route good import/export edge cases.
- [ ] Finish terrain `CanDie`, dead terrain yields, ozone/flood/nuke transformations, `AddAdvance`, `RemoveAdvance`, `TransformAdd`, `TransformRemove`, and terraform advance gates.
- [ ] Finish tile improvement fields: `Class`, `Level`, `ConstructionTiles`, `ProductionCost`, `ProductionTime`, `TerrainEffect.Terrain`, `BonusFood`, `BonusProduction`, `BonusGold`, `MoveCost`, `DefenseBonus`, `HealRate`, `Airport`, `Radar`, `RadarRange`, `VisionRange`, `ListeningPost`, `CanSee`, `CanUpgrade`, `Freight`, `Endgame`, `CantBuildOn`, `Excludes`, `IntBorderRadius`, `SquaredBorderRadius`, `TerraformTerrain`, and pillage/upgrade behavior.
- [ ] Classify terrain/tile-improvement presentation fields (`Icon`, `TilesetIndex`, `HutTilesetIndex*`, `Tooltip`, `Statusbar`, `Sound`, `GLHidden`, `IsCityRuin`, `Column`) or project them safely where public clients need them.

### Governments, sliders, readiness, and happiness/crime

- [ ] Finish government coefficients: `FoodCoef`, `ProductionCoef`, `GoldCoef`, `KnowledgeCoef`, `PollutionCoef`, `PollutionUnhappyCoef`, `CrimeCoef`, `CrimeOffset`, `TooManyCitiesThreshold`, `TooManyCitiesCoefficient`, `EmpireDistanceScale`, `MinEmpireDistance`, `MaxEmpireDistance`, `AtHomeRadius`, `OverseasCoef`, `HomeDefeatCoef`, `OverseasDefeatCoef`, and decay fields.
- [ ] Finish social slider parity: `WorkdayExpectation`, `WagesExpectation`, `RationsExpectation`, positive/negative coefficients, unrest/happiness, production/gold/food effects, AI slider elements, and government caps.
- [ ] Finish readiness parity: `ReadyPeaceCoef`, `ReadyAlertCoef`, `ReadyWarCoef`, `ReadyPeaceHP`, `ReadyAlertHP`, `ReadyWarHP`, `TurnsToNewReadiness`, transition timing, support cost, HP modifier interpolation, and UI/projection state.
- [ ] Finish martial-law/war-discontent/conquest-distress: `MartialLawEffect`, `MartialLawThreshold`, `MaxMartialLawUnits`, `WarDiscontentMaxUnits`, `WarDiscontentPerUnit`, `ConquestDistress`, `ConquestDistressDecay`, fortified/home/overseas exceptions, and wonder/building mitigations.
- [ ] Finish rush-buy/trade/government caps: `BuildingRushModifier`, `UnitRushModifier`, `WonderRushModifier`, `EndGameRushModifier`, `CapitalizationCoefficient`, `InfrastructureCoefficient`, `MaxScienceRate`, `MaxIncomingTrade`, `MaxOutgoingTrade`, and `ProfessionalUnits`.
- [ ] Classify government rank/advice/icon fields (`Rank`, `*Rank`, `HigherRankAdvice`, `SameRankAdvice`, `Icon`, `IsEcotopia`) or wire them to AI/projection where source-backed.

### Diplomacy, trade routes, and empire systems

- [ ] Finish contact, regard, proposals, treaty states, embassies, close-embassy effects, at-war embassy rules, and retail regard drift/AI diplomacy behavior.
- [ ] Finish trade-route creation/pathing/cost, transport-point capacity, caravan/freight/trader units, route cancellation, piracy, war/capture removal, water/international classification, and treasury/yield payouts.
- [ ] Finish franchise/injunction/lawsuit systems: franchise production siphon, injunction timers, lawsuit target removal, CanSue/CanExpel city/unit interactions, and AI/legal projection.
- [ ] Finish city capture/recapture effects: founder identity, conquest distress, trade-route removal, improvements/wonders transfer/destruction, slaves, martial law, and regard/war side effects.

### Pollution, disasters, victory, and endgame

- [ ] Finish retail pollution trigger selection by map size and game setup rather than fixed small-map fallback.
- [ ] Finish pollution accumulation/reduction from population, production, buildings, wonders, advances, nukes, terrain, and world pollution reduction effects.
- [ ] Finish flood and ozone disaster target selection, terrain/dead-tile conversion, city/unit/improvement destruction, pollution history, event payloads, and AI response.
- [ ] Finish all victory paths: bloodlust/conquest, diplomatic/world-union, score/final-turn, Gaia/Solaris endgame, space/launch interactions, dead-player lifecycle, rankings, and tie-breaking.
- [ ] Finish Gaia controller retail details beyond prerequisites: countdown interruption/restart, wonder destruction/capture, endgame object loss, map coverage recalculation, and terminal event sequencing.

### AI, projection, API, persistence, and verification

- [ ] Ensure AI uses retail strategies/build lists/public legal actions for research, production, diplomacy, trade, movement, combat, pollution, wonders, endgame, and recovery without hidden-state shortcuts.
- [ ] Ensure projections expose every legal action safely and suppress illegal/stale fog-memory actions, especially special actions, trade, combat, nukes, space, endgame, and diplomacy.
- [ ] Ensure HTTP/SSE command/event contracts exist and are documented for every executable mechanic.
- [ ] Ensure canonical persistence/restart equality covers every new mechanic and schema field; update fullgame observed hashes and `internal/e2e/freeze/engine-lock.json` after production changes.
- [ ] Add targeted red/green/refactor evidence and source citations for every remaining behavior slice.
- [ ] Run package tests after each slice, freeze guard after production `.go` changes, `git diff --check` before commits, and periodic full `go test ./... -count=1`.
- [ ] Complete a final source-parity review proving every gameplay-affecting retail field/action is executed or safely classified, then mark broad category checkboxes complete.

## Verification and Review

- [x] Run targeted RED/GREEN/REFACTOR evidence for every behavior slice.
- [x] Run package tests for changed packages after each slice.
- [x] Run freeze guard after production `.go` changes.
- [x] Run `git diff --check` before every commit.
- [x] Run long retail AI artifact recording after full-game progress fixes.
- [x] Run full repository validation where feasible before final review.
- [x] Complete and record a final maintainability/source-parity code review.
