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
- [x] Task 085: Retail repeated unit `GovernmentType` gates now accept any listed government in production, projection, and upgrades.
- [x] Task 086: Retail building `CantBuildOnLand` now blocks land-city production in engine, legal projections, and AI placement.
- [x] Task 087: Building `DefaultIcon` and `Description` are source-classified as presentation/UI/SLIC metadata.
- [x] Task 088: Building structural markers `Capitol`, `Brokerage`, `Television`, and `DoubleTelevangelists` are source-verified against existing execution tests.
- [x] Task 089: Wonder `DefaultIcon`, `Description`, and `Movie` are source-classified as presentation/UI metadata.
- [x] Task 090: Advance `Branch`, `Icon`, `GLHidden`, and `KeyAdvanceAdvice` are source-classified; `Age` is split out as gameplay/setup work.
- [x] Task 091: Government rank/advice/icon fields and `IsEcotopia` are source-classified as UI/advice/marker metadata.
- [x] Task 092: Terrain and tile-improvement presentation/editor fields are source-classified as non-gameplay metadata.
- [x] Task 093: Active-defense carrier/enabler gates are source-verified against existing bombard resolution tests.
- [x] Task 094: Advance age metadata, starting-age advance grants, end-age research suppression, and research-completion age transitions are implemented.
- [x] Task 095: Retail `IsFlanker` grouped-combat retargeting is source-verified against existing engine tests.
- [x] Task 096: Retail combat stats and core modifiers are source-verified against engine/rules regression coverage.
- [x] Task 097: Retail grouped combat/city-capture edge unit flags are source-verified; `Civilian` is classified as AI/diplomacy metadata for combat execution.
- [x] Task 098: Advance map-action flags are source-classified/verified against tile-improvement and terraform gates.
- [x] Task 099: Remaining gameplay-bearing advance flags (`PlantNuke`, capitalization/infrastructure, pollution modifiers, `TriggerFeat`) are source-verified.
- [x] Task 100: Field coverage ledger now has a regression guard requiring every parsed retail field to be mentioned.
- [x] Task 101: Production unlock gates now require all runtime prerequisites and reject any listed obsolete advance during queueing and cleanup.
- [x] Task 102: `UpgradeTo` is source-classified as released-data marker/non-retail executable; optional server upgrade command/projection/AI tests are verified.
- [x] Task 103: Special/spy action vector execution, projection, HTTP routes, and public-AI consumption are source-verified.
- [x] Task 104: `SPECIAL_ACTION_MOVE_COST` imports to runtime constants and drives all special-action movement spending.
- [x] Task 105: Movement/fuel/refuel/ZOC/beach/paratrooper/visibility/stealth/global-radar parity is source-verified with regression coverage.
- [x] Task 106: Retail `IsTrader`/`CARAVAN_COEF` behavior now treats trader production as freight capacity, consumes route transport cost, records public used/total freight, and suppresses over-capacity legal/AI route planning.
- [x] Task 107: Retail wonder-granted effective buildings now contribute building economy effects (food/production/commerce/science/gold-per-citizen) through the same city-local effect path without double-counting imported wonder effects.
- [x] Task 108: Retail building FoodVat effects now read actual non-obsolete building effects, and population/happiness fields are source-verified with Cathedral/IsReligious classified as no-extra-mutation markers for released data.
- [x] Task 109: Retail building crime/pollution effects are source-verified: LowerCrime and pollution percent fields execute through effective-building-aware city effects, NuclearPlant cleanup is covered, and absent released PollutionAmount needs no runtime mutation.
- [x] Task 110: Retail building protection effects execute for direct and wonder-granted effective buildings, including nuke/force-field, bio/nano, conversion, and slavery shields.
- [x] Task 111: Retail building military effects execute for direct and effective buildings; Airport is source-classified as a city marker while city/airbase refuel already follows retail source behavior.
- [x] Task 112: Retail effective-building brokerage effects now preserve building-local commerce/brokerage grouping so `IncreaseBrokerages` amplifies wonder-granted brokerages.
- [x] Task 113: Retail numeric wonder effects are source-verified and active-wonder gated for yield, research, happiness, crime, empire size, route/television gold, and flat bonus-gold fields.
- [x] Task 114: Retail global/diplomacy wonder effects execute or project for close/everywhere embassies, at-war embassy edge cases, regard, spies, barbarian protection, and revolt redirection.
- [x] Task 115: Retail production/prohibition wonder effects execute with illegal queue cleanup and reversible active-wonder gating for destroyed/obsolete EliminateNukes wonders.
- [x] Task 116: Retail happiness/readiness/pollution wonder effects execute with expiring full-happiness markers, pollution-unhappiness suppression, and stacked/capped readiness-cost reduction.
- [x] Task 117: Retail visibility/endgame wonder effects execute or project for GlobalRadar, StartGaiaController, BuildingEverywhere, ActualBuildingEverywhere, obsolescence, destruction, capture, and transfer behavior.
- [x] Task 118: Retail terrain EnvRiver modifiers now import and execute for river yield/movement/defense, alongside existing EnvBase/EnvCity city-center selection.
- [x] Task 119: Retail terrain Resources/goods spawning is classified as worldgen input, while canonical tile/city/trade-route goods and NeedsCityGood import/export mechanics execute in gameplay.
- [x] Task 120: Retail terrain transform advance gates, CanDie, and dead-yield metadata import to runtime; terraform requires source remove/target add advances and ozone only kills CanDie terrain.
- [x] Task 121: Retail tile-improvement metadata imports/projects, border-radius forts claim territory, and detector overlays can reveal underwater units; freight/upgrade markers are source-classified.
- [x] Task 122: Retail government economy/crime/pollution/empire/rush/trade/readiness coefficient import/execution is verified; `ProfessionalUnits` is classified as AI/outcom metadata with unit-level professional support markers executing.
- [x] Task 123: Retail social slider expectations, coefficients, economy/happiness effects, AI legal-action projection, and owner-only projection are verified.
- [x] Task 124: Retail readiness, support, HP modifiers, martial law, war discontent, conquest distress, and peace-movement home/overseas decay mechanics are verified.
- [x] Task 125: Retail pollution trigger selection now uses canonical world tile-count map size instead of fixed small-map fallback.
- [x] Task 126: Retail pollution accumulation/reduction, nuclear pollution, trigger history, and flood/ozone disaster outcomes are verified for modeled gameplay paths.
- [x] Task 127: Retail diplomacy/trade/franchise/legal/city-capture interactions are verified across public projections, trade routes, piracy, route removal, capture side effects, and AI legal action paths.

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
- [x] Complete runtime execution for all gameplay-affecting wonder fields.
- [x] Complete runtime execution for all gameplay-affecting building fields.
- [ ] Complete runtime execution for all terrain, worker, public-works, goods, and trade effects.
- [ ] Complete runtime execution for all government, diplomacy, slider, readiness, happiness, crime, and pollution effects.
- [ ] Complete runtime execution for all victory/endgame/space/Gaia/final-turn mechanics.
- [ ] Update the retail field coverage report so no behavior-bearing retail field remains unimplemented or parse-only.

## Detailed Remaining Retail Parity Ledger

This ledger is the authoritative remaining-task checklist. It must be kept current: when a slice is implemented, mark or refine the corresponding checkbox and add the concrete completed task under "Completed Increments" / numbered task history with source evidence.

### Source coverage and classification

- [ ] Re-audit every released `gamedata` catalog field against `ctp2/ctp2_source_code` call sites and classify each as executed gameplay, safely projected marker, presentation-only, AI-only, or unsupported bug-fixed behavior.
- [ ] Update `.sisyphus/evidence/ctp2-full-retail-task-22-field-coverage-inventory.md` so no gameplay-affecting field remains "parse-only" or ambiguously classified.
- [x] Add/maintain regression tests that fail when newly parsed retail fields are omitted from the coverage/classification ledger.
- [ ] Import and execute remaining gameplay constants from `Const.txt` that affect modeled mechanics, including fuel, special-action chances, unrest, and disaster constants; `SPECIAL_ACTION_MOVE_COST` and trade-route `CARAVAN_COEF` path cost are imported/executed.

### Advances and research

- [x] Verify and close all advance unlock gates for unit/building/wonder/terrain-improvement production, including multi-prerequisite edge cases and obsolete-advance removal.
- [x] Complete advance `Age` gameplay/setup parity for start/end age, age transitions, and age-gated initial grants; `Branch`, `Icon`, `GLHidden`, and `KeyAdvanceAdvice` are source-classified as presentation/advice metadata.
- [x] Finish source-backed execution/verification for `DeepOcean`, `Roads`, `Railroads`, `MagLev`, `Tunnels`, and `Transform` map-action gates.
- [x] Verify `PlantNuke`, `Capitalization`, `Infrastructure`, `PollutionProductionModifier`, `PollutionSizeModifier`, and `TriggerFeat` parity across government/advance transitions and persistence.

### Unit catalog fields, movement, support, cargo, and visibility

- [x] Execute retail material support completely: `ShieldHunger`, `NeedsNoSupport`, `IsSpecialForces`, readiness support coefficients, fractional cost rounding, disband ordering, and wonder readiness reductions.
- [x] Source-classify `ProfessionalUnits` government quota/behavior as default-zero AI utility metadata; per-unit professional support/disband behavior is implemented.
- [x] Source-check and close `GoldHunger` and `FoodHunger`: classified as non-executed/default-disabled upkeep metadata for literal retail parity.
- [x] Complete `IsTrader` behavior: trader unit production is modeled as transport-point capacity, route path-cost consumption uses imported `CARAVAN_COEF`, over-capacity routes are rejected/suppressed in public legal actions, and projection-only AI consumes only those public route actions; trader death/removal is represented by capacity abstraction rather than map units.
- [x] Execute retail cargo semantics for `CargoData.MaxCargo`, `CanCarry`, domain/size constraints, engine embark, legal public projection, AI embark planning, transport-destruction cargo fate, and `CargoData.Load`/`CargoData.Unload` presentation-only classification.
- [x] Finish air/space/sea/land/domain movement parity, including `MovementType`, `MaxMovePoints`, `MaxFuel`, `NoFuelThenCrash`, `LossMoveToDmgNone`, non-space fuel costs, crash/damage rules, city/airport/carrier refuel, and hostile terrain immunity; free-space per-turn movement/`SPACE_FUEL_COST` is classified outside the current server map-movement layer.
- [x] Verify `CanBeachAssault`, `Paratrooper`, and `ParatrooperTransport` beyond unload movement: legal target selection, transport requirements, and post-drop movement/combat restrictions.
- [x] Complete `VisionRange`, `VisionClass`, `CanSee`, `IsSubmarine`, `Stealth`, `CanSeeStealth`, `InvisibleForGlobeSat`, global radar, radar/listening-post interactions, and stale-fog legal action suppression.
- [x] Verify `NoZoc`/`IgnoreZOC`/ZOC movement stops and AI pathing behavior against retail.
- [x] Complete unit production restrictions: `BuildingRemovesAPop`, `CantBuild`, `OnlyBuildOne`, repeated `GovernmentType` gates, `NeedsCityGood`, `NoSlaves`, `EnableAdvance`, and `ObsoleteAdvance` are implemented; `UpgradeTo` is source-classified as a released-data marker with optional fail-closed server execution; production-cleanup edge cases are covered. `HasPopAndCanBuild` city pseudo-units are rejected from normal queues, and `NoIndex`/`GLHidden` are presentation/setup metadata.
- [x] Complete upgrade semantics for `UpgradeTo`: no retail source-backed upgrade cost exists; optional server upgrades require target prerequisites/government, spend movement, preserve HP proportionally, reset fuel from target stats, project safe legal actions, are AI-consumable, HTTP-routable through canonical commands, and persist through canonical state.
- [x] Classify all unit sound/icon/sprite fields (`DefaultIcon`, `DefaultSprite`, `Sound*`) as presentation/setup metadata for server gameplay.

### Unit combat and special actions

- [x] Implement `IsFlanker` tactical targeting, including grouped-combat target row/column behavior.
- [x] Source-classify `Explodes` death/combat behavior as presentation-only battle explosion animation metadata.
- [x] Verify all combat stats and modifiers: `Attack`, `Defense`, `ZBRangeAttack`, `BombRounds`, `BombardRange`, `Firepower`, `Armor`, `MaxHP`, `MountedBonus`, `BonusAgainstMounted`, `AttackWoodenShipBonus`, `WoodenShipBonus`, `Defend`, `CanAttack`, `CanBombard`, `CanCounterBombard`, `CityWalls`, `ReduceCityWalls`, and `IgnoreCityWalls`.
- [x] Complete active defense parity: `ActiveDefenseRange`, `ActiveDefenseOnlyWhenCarryingEnablers`, `EnableCarrierDefenses`, interceptor target selection, once-per-turn use/reset, enabler cargo checks, and bombard/counter-bombard RNG ordering.
- [x] Complete `Civilian`, `CantCaptureCity`, `VictoryEnslavement`, `DeathEffectsHappy`, `SingleUse`, and `NuclearAttack` edge cases under grouped combat/city capture; `Civilian` has no separate retail grouped-combat/capture mutation beyond AI/diplomacy/threat classification.
- [x] Complete all spy/special-action vectors field-by-field: `EstablishEmbassy`, `StealTechnology`, `PlantNuke`, `InvestigateCity`, `AssasinateRuler`, `BioTerror`, `NanoTerror`, `Plague`, `ConvertCities`, `CreateFranchise`, `CauseUnhappiness`, `Advertise`, `InciteRevolution`, `IndulgenceSales`, `SlaveRaids`, `SettlerSlaveRaids`, `SlaveUprising`, `UndergroundRailway`, `CanInjoin`, `CanReform`, `CanSue`, `CanExpel`, `CreateParks`, `ThrowParty`, and nested `Chance`/`DeathChance`/`EliteChance`/`EscapeChance`/`Effect`/`Timer`/`Amount` fields are executed; nested `Sound` is classified presentation/audio metadata.
- [x] Ensure every special action has safe public legal-action projection, HTTP route coverage where applicable, canonical event payloads, AI-consumable legality, and no hidden-state shortcut.

### Buildings and city effects

- [x] Finish all building yield/economy effects: `FoodPercent`, `ProductionPercent`, `CommercePercent`, `SciencePercent`, `GoldPerCitizen`, `Upkeep`, and interactions with sliders/government coefficients, including wonder-granted effective buildings.
- [x] Finish population/happiness effects: `HappyInc`, `RaiseMaxPopulation`, `RaiseOvercrowdingLevel`, `FoodVat`, `StarvationProtection`, `LowerPeaceMovement`, and temporary city effects execute; `Cathedral`/`IsReligious` are source-classified as no-extra-mutation released-data markers.
- [x] Finish crime/pollution effects: `LowerCrime`, `PopulationPollutionPercent`, `ProductionPollutionPercent`, `NuclearPlant`, and wonder/effective-building interactions execute; flat `PollutionAmount` is source-classified as absent from released data.
- [x] Finish protection effects: `ProtectFromNukes`, `ProtectFromBioAgents`, `ProtectFromNanoVirus`, `PreventConversion`, `PreventSlavery`, and `ForceField` for both direct buildings and effective buildings.
- [x] Finish military/building effects: `DefendersPercent`, `OffenseBonusLand`, `OffenseBonusWater`, `OffenseBonusAir`, `CityWalls`, `Airport`, and effective-building consequences; `AllowGrunts` is source-classified as non-executed/AI-placeholder metadata.
- [x] Finish structural/build-gate effects: `PrerequisiteBuilding`, `CoastalBuilding`, `CantBuildInSea`, and `CantBuildOnLand` placement gates are implemented; `Capitol`, `Brokerage`, `Television`, and `DoubleTelevangelists` execute with source-backed tests; building `DefaultIcon` and `Description` are presentation/UI metadata.
- [x] Ensure wonder-granted effective buildings participate in every remaining retail `GetEffectiveBuildings()` use: happiness, crime, pollution, protection, city walls/defense/offense, television, brokerage, and production prerequisites; refuel/airport source behavior is city/airbase-based with Airport classified as a marker.

### Wonders

- [x] Finish numeric wonder effects: `IncKnowledgePercent`, `IncreaseFoodAllCities`, `IncreaseProduction`, `IncHappinessEmpire`, `DecCrimePercent`, `DecEmpireSize`, `BonusGold`, `GoldPerWaterTradeRoute`, `GoldPerInternationalTradeRoute`, and `GoldPerTelevision` including obsolescence/destruction.
- [x] Finish global/diplomacy wonder effects: `CloseEmbassies`, `EmbassiesEverywhereEvenAtWar`, `IncreaseRegard`, `SpiesEverywhere`, `ProtectFromBarbarians`, `RevoltingCitiesJoinPlayer`, and relation/war/embassy edge cases.
- [x] Finish production/prohibition wonder effects: `EnableParkRangers`, `FreeSlaves`, `ProhibitSlavers`, `EliminateNukes`, `ReformCities`, `PollutersToParks`, and illegal queue cleanup/reversal edge cases.
- [x] Finish happiness/readiness/pollution wonder effects: `TemporaryFullHappiness`, `NoPollutionUnhappiness`, `ReduceReadinessCost`, and source-backed duration/stacking behavior.
- [x] Finish visibility/endgame wonder effects: `GlobalRadar`, `StartGaiaController`, `BuildingEverywhere`, `ActualBuildingEverywhere`, obsolescence, destruction, capture, and transfer behavior.
- [x] Classify wonder `DefaultIcon`, `Description`, and `Movie` as presentation/UI metadata for server gameplay.

### Terrain, tile improvements, goods, and map rules

- [x] Finish terrain `EnvBase`, `EnvCity`, and `EnvRiver` yields/movement/score/defense, including river overlays and city-center yield selection; score remains presentation/scoring metadata outside modeled tile yield.
- [x] Implement or classify terrain `Resources`, goods spawning, goods trade bonuses, `NeedsCityGood`, and route good import/export edge cases.
- [x] Finish terrain `CanDie`, dead terrain yields, ozone/flood/nuke transformations, `AddAdvance`, `RemoveAdvance`, `TransformAdd`, `TransformRemove`, and terraform advance gates.
- [x] Finish tile improvement fields: `Class`, `Level`, `ConstructionTiles`, `ProductionCost`, `ProductionTime`, `TerrainEffect.Terrain`, `BonusFood`, `BonusProduction`, `BonusGold`, `MoveCost`, `DefenseBonus`, `HealRate`, `Airport`, `Radar`, `RadarRange`, `VisionRange`, `ListeningPost`, `CanSee`, `CanUpgrade`, `Freight`, `Endgame`, `CantBuildOn`, `Excludes`, `IntBorderRadius`, `SquaredBorderRadius`, `TerraformTerrain`, and pillage/upgrade behavior; retail source exposes no tile-improvement upgrade command, so `CanUpgrade` remains a projected marker.
- [x] Classify terrain/tile-improvement presentation fields (`Icon`, `TilesetIndex`, `HutTilesetIndex*`, `Tooltip`, `Statusbar`, `Sound`, `GLHidden`, `IsCityRuin`, `Column`) as presentation/editor/Great-Library/control-panel metadata for server gameplay.

### Governments, sliders, readiness, and happiness/crime

- [x] Finish government coefficients: `FoodCoef`, `ProductionCoef`, `GoldCoef`, `KnowledgeCoef`, `PollutionCoef`, `PollutionUnhappyCoef`, `CrimeCoef`, `CrimeOffset`, `TooManyCitiesThreshold`, `TooManyCitiesCoefficient`, `EmpireDistanceScale`, `MinEmpireDistance`, `MaxEmpireDistance`, `AtHomeRadius`, `OverseasCoef`, `HomeDefeatCoef`, `OverseasDefeatCoef`, and decay fields.
- [x] Finish social slider parity: `WorkdayExpectation`, `WagesExpectation`, `RationsExpectation`, positive/negative coefficients, unrest/happiness, production/gold/food effects, AI slider elements, and government caps.
- [x] Finish readiness parity: `ReadyPeaceCoef`, `ReadyAlertCoef`, `ReadyWarCoef`, `ReadyPeaceHP`, `ReadyAlertHP`, `ReadyWarHP`, `TurnsToNewReadiness`, transition timing, support cost, HP modifier interpolation, and UI/projection state.
- [x] Finish martial-law/war-discontent/conquest-distress: `MartialLawEffect`, `MartialLawThreshold`, `MaxMartialLawUnits`, `WarDiscontentMaxUnits`, `WarDiscontentPerUnit`, `ConquestDistress`, `ConquestDistressDecay`, fortified/home/overseas exceptions, and wonder/building mitigations.
- [x] Finish rush-buy/trade/government caps: `BuildingRushModifier`, `UnitRushModifier`, `WonderRushModifier`, `EndGameRushModifier`, `CapitalizationCoefficient`, `InfrastructureCoefficient`, `MaxScienceRate`, `MaxIncomingTrade`, `MaxOutgoingTrade`, and `ProfessionalUnits` classification.
- [x] Classify government rank/advice/icon fields (`Rank`, `*Rank`, `HigherRankAdvice`, `SameRankAdvice`, `Icon`, `IsEcotopia`) as UI/advice/marker metadata for server gameplay.

### Diplomacy, trade routes, and empire systems

- [x] Finish contact, regard, proposals, treaty states, embassies, close-embassy effects, at-war embassy rules, and retail regard drift/AI diplomacy behavior.
- [x] Finish trade-route creation/pathing/cost, transport-point capacity, caravan/freight/trader units, route cancellation, piracy, war/capture removal, water/international classification, and treasury/yield payouts.
- [x] Finish franchise/injunction/lawsuit systems: franchise production siphon, injunction timers, lawsuit target removal, CanSue/CanExpel city/unit interactions, and AI/legal projection.
- [x] Finish city capture/recapture effects: founder identity, conquest distress, trade-route removal, improvements/wonders transfer/destruction, slaves, martial law, and regard/war side effects.

### Pollution, disasters, victory, and endgame

- [x] Finish retail pollution trigger selection by map size and game setup rather than fixed small-map fallback.
- [x] Finish pollution accumulation/reduction from population, production, buildings, wonders, advances, nukes, terrain, and world pollution reduction effects.
- [x] Finish flood and ozone disaster target selection, terrain/dead-tile conversion, city/unit/improvement destruction, pollution history, event payloads, and AI response.
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
