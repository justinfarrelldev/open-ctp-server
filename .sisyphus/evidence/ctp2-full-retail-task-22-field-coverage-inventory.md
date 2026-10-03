# CTP2 full-retail parity task 22 field coverage inventory

Scope: inventory every unique released-game gameplay/catalog field currently parsed from the retail CTP2 data files, and record high-level runtime/execution coverage status. This is an inventory only; it is not a full-parity claim.

Source data root:
`ctp2/ctp2_released_game/ctp2_data/default/gamedata`

Regression guard:
- `internal/gameplay/rules/retail_field_inventory_test.go`
- `TestRetailFieldInventorySnapshot` verifies the released catalog field counts and sentinel behavior-affecting fields.
- `TestRetailFieldCoverageLedgerMentionsEveryParsedField` verifies every parsed released catalog field is mentioned in this coverage/classification ledger.

## Summary counts

| Catalog | Records | Unique fields, including nested fields | Current coverage summary |
| --- | ---: | ---: | --- |
| Units | 79 | 194 | Core cost/prereq/obsolete/combat/move/fuel/ranged/bombard plus retail veteran attack odds, `SpaceLaunch`, `NuclearAttack`, `ThrowParty`, `Advertise`, `CanBeGifted`, `CanPirate`, `DeathEffectsHappy`, `InvisibleForGlobeSat`, `IsTrader` freight capacity, cargo capacity/size-class embark gates, and beach-assault unload movement behavior are partially projected/executed. Remaining support/economy edge effects and sound/UI fields are either pending or classified below. |
| Advances | 107 | 19 | Cost/prerequisites, research unlocks, production obsolescence, plant-nuke unlocks, pollution modifiers, capitalization/infrastructure unlocks, terrain-improvement transport/terraform gates, and advance-triggered feats are executed; Age/Branch/Icon/GLHidden/advice are presentation metadata. |
| Feats | 14 | 13 | Advance-triggered feat effects for production, science, happiness, boat movement, hit points, and city-wall reduction are imported as timed runtime effects; building thresholds, sailed-around-world, conquered-by-force, and founder city-recapture triggers now execute as timed nation effects with retail one-time achievement gating. |
| Wonders | 30 | 39 | Cost/prereq/obsolete metadata plus selected effects including `BuildingEverywhere`/`ActualBuildingEverywhere` effective buildings, `IncreaseBoatMovement`, `IncreaseHp`, `IncreaseScientists`, `IncreaseBrokerages`, `IncreaseRegard`, treasury-level gold (`BonusGold`, route gold, television gold), `EnableParkRangers`, `EliminateNukes`, `FreeSlaves`, and `StartGaiaController` countdown requirements are projected/executed; remaining global effects require more execution coverage. |
| Buildings | 52 | 42 | Cost/upkeep/prereq/obsolete, food/production/commerce/science percent effects, `GoldPerCitizen`, population/happiness/growth fields, crime/pollution fields, protection fields, combat/defense/offense/wall fields, effective-building economy/crime/pollution/protection/military/brokerage propagation, television/brokerage trade-gold hooks, `DoubleTelevangelists` conversion gating, `Airport` marker/refuel classification, and `NuclearPlant` production/blocking/cleanup under `EliminateNukes` are projected/executed or source-classified for all released gameplay-bearing building fields. |
| Terrain improvements | 40 | 36 | Class, construction cost/time, movement/yield effects, CantBuildOn, Excludes, vision/radar/listening-post projection, heal, airport refuel, terraform, and Gaia endgame processing-tower coverage are partially supported. Remaining edge behavior and presentation metadata require closure. |
| Terrain | 26 | 43 | EnvBase yields are used for city/production yield. EnvCity/EnvRiver, resources, transforms, hut/tile-set, death and advance gates remain pending. |
| Governments | 12 | 69 | EnableAdvance and economy/research coefficients are partially applied. Readiness, martial law, support, pollution, ranks/advice, rations/wages/workday, trade caps, rush modifiers, and war discontent remain pending. |
| Pollution trigger lists | 4 | 1 | Trigger levels/disaster types are projected and phase/history integration is wired. Full disaster map/city/unit outcomes remain pending. |

## Full field inventory

<details><summary>Units (194)</summary>

ActiveDefenseOnlyWhenCarryingEnablers, ActiveDefenseRange, Advertise, Armor, AssasinateRuler, AssasinateRuler.Chance, AssasinateRuler.DeathChance, AssasinateRuler.EliteChance, Attack, AttackWoodenShipBonus, BioTerror, BioTerror.Chance, BioTerror.Effect, BioTerror.Sound, BombRounds, BombardRange, BonusAgainstMounted, BuildingRemovesAPop, CanAttack, CanBeExpelled, CanBeGifted, CanBeSued, CanBeachAssault, CanBombard, CanCarry, CanCounterBombard, CanEntrench, CanExpel, CanInjoin, CanInjoin.Effect, CanInjoin.Sound, CanPillage, CanPirate, CanReform, CanReform.Effect, CanReform.Sound, CanSee, CanSoothsay, CanSoothsay.Effect, CanSoothsay.Sound, CanSue, CantBuild, CantCaptureCity, CargoData, CargoData.Load, CargoData.MaxCargo, CargoData.Unload, Category, CauseUnhappiness, CauseUnhappiness.Amount, CauseUnhappiness.Chance, CauseUnhappiness.Effect, CauseUnhappiness.Sound, CauseUnhappiness.Timer, CityGrowthCoefficient, Civilian, ConvertCities, ConvertCities.Chance, ConvertCities.DeathChance, ConvertCities.Effect, ConvertCities.Sound, CreateFranchise, CreateFranchise.Chance, CreateFranchise.Effect, CreateFranchise.Sound, CreateParks, CreateParks.Effect, CreateParks.Sound, DeathEffectsHappy, DefaultIcon, DefaultSprite, Defend, DefendAgainstSpies, Defense, Description, EnableAdvance, EnableCarrierDefenses, EstablishEmbassy, ExertsMartialLaw, Explodes, Firepower, FoodHunger, GLHidden, GoldHunger, GovernmentType, HasPopAndCanBuild, IgnoreZOC, ImmuneToHostileTerrain, InciteRevolution, InciteRevolution.Chance, InciteRevolution.DeathChance, InciteRevolution.Effect, InciteRevolution.EliteChance, InciteRevolution.Sound, IndulgenceSales, IndulgenceSales.Effect, IndulgenceSales.Sound, InvestigateCity, InvestigateCity.Chance, InvestigateCity.DeathChance, InvestigateCity.EliteChance, InvisibleForGlobeSat, IsFlanker, IsMounted, IsSpecialForces, IsSubmarine, IsTelevangelist, IsTrader, LossMoveToDmgNone, MaxFuel, MaxHP, MaxMovePoints, MountedBonus, MovementType, NanoTerror, NanoTerror.Chance, NanoTerror.Effect, NanoTerror.Sound, NeedsCityGood, NeedsNoSupport, NoFuelThenCrash, NoIndex, NoSlaves, NoZoc, NuclearAttack, NuclearAttack.Effect, NuclearAttack.Sound, ObsoleteAdvance, OnlyBuildOne, Paratrooper, ParatrooperTransport, Plague, Plague.Chance, Plague.Effect, Plague.Sound, PlantNuke, PlantNuke.Chance, PlantNuke.Effect, PlantNuke.EscapeChance, PlantNuke.Sound, PowerPoints, Revolution, Revolution.Effect, Revolution.Sound, Settle, SettleBuilding, SettleCityType, SettleSize, SettlerSlaveRaids, SettlerSlaveRaids.Effect, SettlerSlaveRaids.Sound, ShieldCost, ShieldHunger, SingleUse, Size, SlaveRaids, SlaveRaids.Amount, SlaveRaids.Chance, SlaveRaids.DeathChance, SlaveRaids.Effect, SlaveRaids.Sound, SlaveRaids.Timer, SlaveUprising, SlaveUprising.Effect, SlaveUprising.Sound, SoundAcknowledge, SoundAttack, SoundCantMove, SoundDeath, SoundMove, SoundSelect1, SoundSelect2, SoundVictory, SoundWork, SpaceLaunch, StealTechnology, StealTechnology.DeathChance, StealTechnology.Effect, StealTechnology.RandomChance, StealTechnology.Sound, StealTechnology.SpecificChance, ThrowParty, UndergroundRailway, UndergroundRailway.Chance, UndergroundRailway.DeathChance, UndergroundRailway.Effect, UndergroundRailway.Sound, UpgradeTo, VictoryEnslavement, VisionClass, VisionRange, WoodenShip, WoodenShipBonus, ZBRangeAttack

</details>

<details><summary>Advances (19)</summary>

Age, Branch, Capitalization, Cost, DeepOcean, GLHidden, Icon, Infrastructure, KeyAdvanceAdvice, MagLev, PlantNuke, PollutionProductionModifier, PollutionSizeModifier, Prerequisites, Railroads, Roads, Transform, TriggerFeat, Tunnels

</details>

<details><summary>Feats (13)</summary>

Building, Building.Building, Description, Duration, EffectBoatMovement, EffectIncreaseCommerce, EffectIncreaseHappiness, EffectIncreaseHitPoints, EffectIncreaseProduction, EffectIncreaseScience, EffectReduceCityWalls, MinimumSizeOfCiv, SlicMessage

</details>

<details><summary>Wonders (39)</summary>

ActualBuildingEverywhere, BonusGold, BuildingEverywhere, CloseEmbassies, DecCrimePercent, DecEmpireSize, DefaultIcon, Description, EliminateNukes, EmbassiesEverywhereEvenAtWar, EnableAdvance, EnableParkRangers, FreeSlaves, GlobalRadar, GoldPerInternationalTradeRoute, GoldPerTelevision, GoldPerWaterTradeRoute, IncHappinessEmpire, IncKnowledgePercent, IncreaseBoatMovement, IncreaseBrokerages, IncreaseFoodAllCities, IncreaseHp, IncreaseProduction, IncreaseRegard, IncreaseScientists, Movie, ObsoleteAdvance, PollutersToParks, PreventConversion, ProductionCost, ProhibitSlavers, ProtectFromBarbarians, ReduceReadinessCost, ReformCities, RevoltingCitiesJoinPlayer, SpiesEverywhere, StartGaiaController, TemporaryFullHappiness

</details>

<details><summary>Buildings (42)</summary>

Airport, AllowGrunts, Brokerage, CantBuildInSea, Capitol, Cathedral, CityWalls, CoastalBuilding, CommercePercent, DefaultIcon, DefendersPercent, Description, DoubleTelevangelists, EnableAdvance, FoodPercent, FoodVat, ForceField, GoldPerCitizen, HappyInc, IsReligious, LowerCrime, LowerPeaceMovement, NuclearPlant, OffenseBonusAir, OffenseBonusLand, OffenseBonusWater, PopulationPollutionPercent, PrerequisiteBuilding, PreventConversion, PreventSlavery, ProductionCost, ProductionPercent, ProductionPollutionPercent, ProtectFromBioAgents, ProtectFromNanoVirus, ProtectFromNukes, RaiseMaxPopulation, RaiseOvercrowdingLevel, SciencePercent, StarvationProtection, Television, Upkeep

</details>

<details><summary>Terrain improvements (36)</summary>

CanSee, CantBuildOn, Class, Column, ConstructionTiles, Excludes, GLHidden, Icon, IntBorderRadius, IsCityRuin, Level, Sound, SquaredBorderRadius, Statusbar, TerraformTerrain, TerrainEffect, TerrainEffect.Airport, TerrainEffect.BonusFood, TerrainEffect.BonusGold, TerrainEffect.BonusProduction, TerrainEffect.CanUpgrade, TerrainEffect.DefenseBonus, TerrainEffect.EnableAdvance, TerrainEffect.Endgame, TerrainEffect.Freight, TerrainEffect.HealRate, TerrainEffect.ListeningPost, TerrainEffect.MoveCost, TerrainEffect.ProductionCost, TerrainEffect.ProductionTime, TerrainEffect.Radar, TerrainEffect.RadarRange, TerrainEffect.Terrain, TerrainEffect.TilesetIndex, TerrainEffect.VisionRange, Tooltip

</details>

<details><summary>Terrain (43)</summary>

AddAdvance, CanDie, EnvBase, EnvBase.DeadFood, EnvBase.DeadGold, EnvBase.DeadShield, EnvBase.Defense, EnvBase.Food, EnvBase.Freight, EnvBase.Gold, EnvBase.Movement, EnvBase.Score, EnvBase.Shield, EnvCity, EnvCity.CanUpgrade, EnvCity.Food, EnvCity.Freight, EnvCity.Gold, EnvCity.Movement, EnvCity.Score, EnvCity.Shield, EnvRiver, EnvRiver.Food, EnvRiver.Freight, EnvRiver.Gold, EnvRiver.Movement, EnvRiver.Score, EnvRiver.Shield, GLHidden, HutTilesetIndexA, HutTilesetIndexB, Icon, InternalType, MovementType, RemoveAdvance, Resources, TilesetIndex, TransformAdd, TransformAdd.Materials, TransformAdd.Time, TransformRemove, TransformRemove.Materials, TransformRemove.Time

</details>

<details><summary>Governments (69)</summary>

AtHomeRadius, BuildingRushModifier, CapitalizationCoefficient, CommerceRank, ConquestDistress, ConquestDistressDecay, CrimeCoef, CrimeOffset, DefenseCoef, EmpireDistanceScale, EnableAdvance, EndGameRushModifier, FoodCoef, GoldCoef, GoldRank, GrowthRank, HigherRankAdvice, HomeDefeatCoef, HomeDefeatDecay, Icon, InfrastructureCoefficient, IsEcotopia, KnowledgeCoef, LoyaltyRank, MartialLawEffect, MartialLawRank, MartialLawThreshold, MaxEmpireDistance, MaxIncomingTrade, MaxMartialLawUnits, MaxOutgoingTrade, MaxScienceRate, MilitaryRank, MinEmpireDistance, NegativeRationsCoef, NegativeWagesCoef, NegativeWorkdayCoef, OverseasCoef, OverseasDefeatCoef, OverseasDefeatDecay, PollutionCoef, PollutionRank, PollutionUnhappyCoef, PositiveRationsCoef, PositiveWagesCoef, PositiveWorkdayCoef, ProductionCoef, ProductionRank, ProfessionalUnits, Rank, RationsExpectation, ReadyAlertCoef, ReadyAlertHP, ReadyPeaceCoef, ReadyPeaceHP, ReadyWarCoef, ReadyWarHP, SameRankAdvice, ScienceRank, SupportCoef, TooManyCitiesCoefficient, TooManyCitiesThreshold, TurnsToNewReadiness, UnitRushModifier, WagesExpectation, WarDiscontentMaxUnits, WarDiscontentPerUnit, WonderRushModifier, WorkdayExpectation

</details>

<details><summary>Pollution trigger lists (1)</summary>

POLLUTION_TRIGGER

</details>

Validation commands:
- `go test ./internal/gameplay/rules -run RetailFieldInventorySnapshot -count=1`

Task 040 update:
- Retail `ThrowParty` is no longer parse-only: canonical units with the imported `throw_party` marker can execute the reception action, spend gold/movement, apply timed diplomacy regard, project legal actions, and submit through `/throw-party`.

Task 041 update:
- Retail wonder `IncreaseRegard` is no longer parse-only for diplomacy projections: contacted foreign players now see positive/friendly regard toward active owners of `IncreaseRegard` wonders.

Task 042 update:
- Retail unit `Advertise` is no longer parse-only: advertiser units project/execute a dedicated `advertise` action and apply the nested `CauseUnhappiness` amount/timer/chance mechanics through HTTP and canonical engine commands.

Task 043 update:
- Retail wonder `EnableParkRangers` now uses the global built-wonder semantics to unlock `CreateParks` unit production for all players once any active Eden Project-style wonder exists.

Task 044 update:
- Retail building `DoubleTelevangelists` now gates whether televangelist-converted cities use the enhanced televangelist conversion percent or fall back to the cleric conversion percent.

Task 045 update:
- Retail building `NuclearPlant` production now respects the global `EliminateNukes` wonder gate and is rejected after any active nuke-elimination wonder exists.

Task 046 update:
- Retail wonder `EliminateNukes` now also destroys existing `NuclearPlant` buildings and removes queued nuclear-plant items from every city, in addition to removing nuclear units.

Task 047 update:
- Retail wonder `FreeSlaves` now uses the global built-wonder semantics to block slavery unit production, including slave-uprising units, for all players once any active emancipation-style wonder exists.

Task 048 update:
- Retail wonder `FreeSlaves` emancipation now removes queued slave raid, settler slave raid, and slave-uprising unit items via the same post-emancipation illegal-item cleanup retail runs.

Task 049 update:
- Retail wonder `EliminateNukes` now globally blocks nuclear-attack unit production from active wonder state, not only from nation markers written during wonder completion.

Task 050 update:
- Retail wonder `EliminateNukes` now removes queued nuclear-attack unit items during defuser cleanup, matching retail `RemoveIllegalItems()` after nukes are globally eliminated.

Task 051 update:
- Retail unit `InvisibleForGlobeSat` now affects projection: such units are hidden from global-radar-only visibility but still appear under normal line-of-sight visibility.

Task 052 update:
- Retail unit `CanBeachAssault`/paratrooper unload semantics now zero movement for normal debarked cargo while preserving post-entry movement for legal beach-assault/paratrooper unloads and air cargo.

Task 053 update:
- Retail unit `CanBeGifted` is no longer marker-only: giftable units expose safe public gift actions for known diplomacy targets, execute ownership transfer at the giver capitol, and submit through `/gift-unit`.

Task 054 update:
- Retail unit `DeathEffectsHappy` now gates combat casualty peace-movement accounting; ordinary combat deaths no longer increment morale/peace-movement lost-unit counters.

Task 055 update:
- Retail wonder gold fields (`BonusGold`, `GoldPerWaterTradeRoute`, `GoldPerInternationalTradeRoute`, and `GoldPerTelevision`) now pay owner treasury directly at begin turn rather than changing per-city yields.

Task 056 update:
- Retail unit `CanPirate` is no longer marker-only: pirate units can set a visible foreign-source route's pirating army and receive route-value piracy gold using the retail piracy waste coefficient.

Task 057 update:
- Retail `StartGaiaController`/endgame object prerequisites now gate countdown/victory on Gaia computer count, power satellite count, processing tower count, and processing-tower map coverage.

Task 058 update:
- Retail `VETERAN_COEF` from `Const.txt` now increases veteran melee attack and ranged bombard/active-defense/counter-bombard hit strength using the released 50% default.

Task 059 update:
- Retail wonder `BuildingEverywhere` and `ActualBuildingEverywhere` now project effective-building markers so active owner wonders satisfy city `PrerequisiteBuilding` checks, matching retail `GetEffectiveBuildings()`.

Task 060 update:
- Retail `GoldPerTelevision` wonder income now checks television effects through effective city buildings, including wonder-granted building effects.

Task 061 update:
- Retail city capture now transfers built wonders in the captured city to the conqueror and reapplies owner-facing wonder add effects such as hit points and effective-building ownership.

Task 062 update:
- Retail city capture now destroys captured `Capitol` buildings before owner transfer, matching `CityData::DestroyCapitol()`.

Task 063 update:
- Retail city capture now clears owner-specific city effects (`Franchise`, `Injunction`, `Watchful`) and timed happiness/special effects, matching `CityData::ResetCityOwner()` cleanup.

Task 064 update:
- Retail `BOMBARD_DESTROY_BUILDING_CHANCE` and `BOMBARD_KILL_POP_CHANCE` now import/project into runtime constants and drive undefended-city bombard building/population damage; public legal actions include only non-stale visible city bombard targets.

Task 065 update:
- Retail city bombard building destruction now revalidates the city's build queue and removes items that became illegal, matching `DestroyRandomBuilding()->RemoveIllegalItems(TRUE)`.

Task 066 update:
- Retail undefended-city bombard now subtracts wall defense from actual city wall buildings only, matching `GetImprovements()`-based `buildingutil_GetCityWallsDefense`.

Task 067 update:
- Retail `IsFlanker` now affects melee row targeting: non-flankers do not retarget from empty same rows, while flankers search nearest occupied compatible rows.

Task 068 update:
- Retail ranged combat now prefers same-row and nearest-row tactical targets before fallback, matching `FindTargetForRangedAttackFrom`.

Task 069 update:
- Retail bombard hostilities now set war state and remove combatant trade routes, aligning bombard with retail `VerifyAttack` hostile/enemy flow.

Task 070 update:
- Retail `LossMoveToDmgTwo` now imports as `loss_move_to_dmg_two` and applies the minimum-two damaged movement refresh rule from `UnitData::ResetMovement`.

Task 071 update:
- Retail counter-bombard now iterates every living defender on the target tile and applies each defender's own `CanCounterBombard`/`CanBombard` rules, matching `ArmyData::Bombard`.

Task 072 update:
- Retail counter-bombard now checks the defending counter-bombarder's own bombard range to the original bombarder, matching `CanCounterBombard()->CanBombard()`.

Task 073 update:
- Retail active-defense range now uses interceptor-to-bombarder distance, matching `ArmyData::GetActiveDefenders` distance checks against the attacking unit.

Task 074 update:
- Retail `Explodes` is classified as presentation-only battle explosion animation metadata: the gameplay layer creates battle events and the UI gates the visual effect on `GetExplodes`, with no gameplay state mutation tied to the field.

Task 075 update:
- Retail `HasPopAndCanBuild` unit records are city pseudo-unit types and are rejected from normal city production queueing, matching retail `C3UnitDB` and queue UI filters.

Task 076 update:
- Retail active-defense interceptors now use canonical once-per-turn state, skip after first active-defense attempt, and clear that state during turn refresh, matching `k_UDF_USED_ACTIVE_DEFENSE`.

Task 077 update:
- Retail active-defense interceptors now resolve in descending attack order with deterministic ID tie-breaks, matching `ArmyData::GetActiveDefenders` insertion ordering.

Task 078 update:
- `FoodHunger` is parsed/displayed but default data is all zero and no gameplay source consumes it; `GoldHunger` is absent from the original source parser and marked by released data comments as an optional Apolyton modification, so both are classified non-executed/default-disabled for literal retail gameplay.

Task 079 update:
- Government `ProfessionalUnits` is default-zero AI utility metadata in retail data/source; gameplay support exemption is driven by per-unit `IsProfessional`, which the server already executes in support/disband calculations.

Task 080 update:
- Retail material support now retains fractional per-unit readiness costs and floors only after summing, including over-budget disband comparisons, matching `MilitaryReadiness::m_cost` and `Player` rounded-cost logic.

Task 081 update:
- Building `AllowGrunts` is parsed and exposed to AI/outcom, but retail `GetCityGruntOutput` returns zero, so it is classified as non-executed/AI-placeholder metadata.

Task 082 update:
- Unit `NoIndex`, `GLHidden`, `DefaultIcon`, `DefaultSprite`, and `Sound*` fields are classified as presentation/setup metadata: setup lists, Great Library visibility, sprite actors, icons, and sound IDs only.

Task 083 update:
- Retail `CanCarry`, `CargoData.MaxCargo`, and `Size` are no longer broad-domain-only: runtime imports exact `Small|Med|Large` + `Land|Air|Water|Space` carry classes, canonical/visible units preserve size metadata, and engine/legal/AI embark filtering requires exact retail-compatible cargo class matching when class data is present.

Task 084 update:
- Nested `CargoData.Load`/`CargoData.Unload` are classified as presentation-only/unreferenced sound metadata: they parse as Sound DB records and expose `SoundRecord` accessors, but retail boarding/unload audio uses top-level `SoundLoad`/`SoundUnload`; transport death cargo fate is verified against existing recursive cargo removal.

Task 085 update:
- Repeated unit `GovernmentType` gates are now preserved as `RequiredGovernmentIDs`; production queue legality, visibility legal-action projection, and upgrade legality accept any listed government and reject unlisted governments, matching retail's `GetNumGovernmentType()` loop.

Task 086 update:
- Building `CantBuildOnLand` is now imported/executed with the same city-tile water/land placement gate as retail `CityData::CanBuildImprovement`: engine production, visible legal actions, and projection-only AI all reject land-city production for such buildings.

Task 087 update:
- Building `DefaultIcon` and `Description` are classified as presentation/UI/SLIC metadata: retail parses them as icon/string IDs and uses them for queue/scenario-editor/Great-Library style presentation, not gameplay state mutation or legality.

Task 088 update:
- Building structural markers `Capitol`, `Brokerage`, `Television`, and `DoubleTelevangelists` are source-verified as covered by existing runtime projection/execution tests: capture/bombard/capital effects, brokerage-commerce wonder amplification, television wonder income/effective buildings, and televangelist conversion gating.

Task 089 update:
- Wonder `DefaultIcon`, `Description`, and `Movie` are classified as presentation/UI metadata: retail parses them as icon/string/movie DB references and consumes them in wonder movie, Great Library, and queue display code only.

Task 090 update:
- Advance `Branch`, `Icon`, and `GLHidden` are classified as presentation/setup metadata, and `KeyAdvanceAdvice` as diplomacy/AI advice text metadata. Advance `Age` is explicitly not classified away because source uses it for start/end-age setup and age-transition gameplay.

Task 091 update:
- Government rank/advice/icon fields (`Rank`, `*Rank`, `HigherRankAdvice`, `SameRankAdvice`, `Icon`) are classified as UI/advice metadata, and `IsEcotopia` as non-executed marker metadata, with no server gameplay state mutation required for literal retail parity.

Task 092 update:
- Terrain/tile-improvement presentation fields (`Icon`, `TilesetIndex`, `HutTilesetIndex*`, `Tooltip`, `Statusbar`, `Sound`, `GLHidden`, `IsCityRuin`, `Column`) are classified as presentation/editor/Great-Library/control-panel metadata; no server gameplay execution is required.

Task 094 update:
- Advance `Age` now imports through `age.txt`/`Advance.txt`, projects to runtime rules, grants earlier-age advances for configured starting age, suppresses research beyond configured ending age in engine/projection, advances `Nation.CurrentAgeID` on the retail `NeedAdvances` threshold, and seeds eligible earlier-age non-capitol city buildings at founding.

Task 095 update:
- Unit `IsFlanker` is executed in grouped combat targeting: non-flankers do not retarget inactive rows, while flankers scan adjacent rows in the same column for the nearest active defender.

Task 096 update:
- Core combat stats/modifiers (`Attack`, `Defense`, `ZBRangeAttack`, `BombRounds`, `BombardRange`, `Firepower`, `Armor`, `MaxHP`, `MountedBonus`, `BonusAgainstMounted`, `AttackWoodenShipBonus`, `WoodenShipBonus`, `Defend`, `CanAttack`, `CanBombard`, `CanCounterBombard`, `CityWalls`, `ReduceCityWalls`, `IgnoreCityWalls`) are source-verified as executed by rules projection and engine combat/bombard tests.

Task 097 update:
- Grouped-combat/city-capture edge fields (`CantCaptureCity`, `VictoryEnslavement`, `DeathEffectsHappy`, `SingleUse`, `NuclearAttack`) are source-verified as executed; `Civilian` is preserved for AI/diplomacy/threat classification and has no separate retail grouped-combat/capture mutation.

Task 098 update:
- Advance map-action flags are closed: `DeepOcean` is presentation/map-visibility refresh metadata; `Transform` gates the retail terraform UI tab while actual server action legality uses terrain transform data and tile-improvement `EnableAdvance`; `Roads`, `Railroads`, `MagLev`, and `Tunnels` have no non-parser source call sites and released data gates their improvements through tile-improvement advances.

Task 099 update:
- Remaining gameplay-bearing advance fields (`PlantNuke`, `Capitalization`, `Infrastructure`, `PollutionSizeModifier`, `PollutionProductionModifier`, `TriggerFeat`) are source-verified as runtime/engine/projection executed.

Task 100 update:
- The field coverage inventory now has a regression test that fails if any parsed released-game catalog field is omitted from this ledger.

Task 101 update:
- Production unlock/obsolete gates are closed for queueing and cleanup: runtime production now requires all prerequisites and rejects any listed obsolete advance/wonder for unit/building/wonder production, matching retail's repeated prerequisite/obsolete loops.

Task 102 update:
- Unit `UpgradeTo` is source-classified as a released-data marker not parsed/executed by the retail source; the server's optional upgrade command remains fail-closed and covered for prerequisites, government gates, legal projection, AI use, and deterministic stat/HP/fuel/movement carryover.

Task 103 update:
- All imported special/spy action vectors are source-verified as executed/projection-covered (`EstablishEmbassy`, `StealTechnology`, `PlantNuke`, `InvestigateCity`, `AssasinateRuler`, `BioTerror`, `NanoTerror`, `Plague`, `ConvertCities`, `CreateFranchise`, `CauseUnhappiness`, `Advertise`, `InciteRevolution`, `IndulgenceSales`, `SlaveRaids`, `SettlerSlaveRaids`, `SlaveUprising`, `UndergroundRailway`, `CanInjoin`, `CanReform`, `CanSue`, `CanExpel`, `CreateParks`, `ThrowParty`); nested sound fields remain presentation/audio metadata.

Task 104 update:
- `Const.txt` `SPECIAL_ACTION_MOVE_COST` is imported/projected and now drives all special-action movement spending (including piracy), replacing the previous synthetic one-point fallback when retail runtime constants are pinned.

Task 105 update:
- Movement/visibility fields are closed: `MovementType`, `MaxMovePoints`, `MaxFuel`, `NoFuelThenCrash`, `LossMoveToDmgNone`, `LossMoveToDmgTwo`, `ImmuneToHostileTerrain`, `NoZoc`/`IgnoreZOC`, `CanBeachAssault`, `Paratrooper`, `ParatrooperTransport`, `VisionRange`, `VisionClass`, `CanSee`, `IsSubmarine`, `Stealth`, `CanSeeStealth`, `InvisibleForGlobeSat`, global radar, tile-improvement radar/listening-post/vision, and stale-fog action suppression are executed or source-classified where outside the current server map layer.

Task 106 update:
- Unit `IsTrader` is no longer marker-only: producing a retail trader unit adds owner freight/transport capacity from the unit's imported `MaxMovePoints`/movement value without placing a normal map unit, matching retail's separate `m_traderUnits`/freight pool. Trade-route creation records and consumes route transport cost derived from imported `Const.txt` `CARAVAN_COEF`, exposes public used/total freight in projections, rejects over-capacity route creation, and suppresses over-capacity public legal actions consumed by AI.

Task 107 update:
- Building economy fields (`FoodPercent`, `ProductionPercent`, `CommercePercent`, `SciencePercent`, `GoldPerCitizen`, and `Upkeep`) are source-verified as executed through city yield/research/maintenance paths. Wonder-granted `BuildingEverywhere`/`ActualBuildingEverywhere` effective buildings now pull the pinned runtime building effects into the city-local effect path without double-counting effects already embedded on imported retail wonders.

Task 108 update:
- Population/happiness building fields are closed: `HappyInc`, `RaiseMaxPopulation`, `RaiseOvercrowdingLevel`, `FoodVat`, `StarvationProtection`, `LowerPeaceMovement`, and temporary full-happiness effects execute in the city lifecycle; `Cathedral`/`IsReligious` are source-classified as no-extra-mutation markers for released retail data because `IncreaseCathedrals` is absent and retail `IsReligious` subtracts zero.

Task 109 update:
- Building crime/pollution fields are closed: `LowerCrime`, `PopulationPollutionPercent`, and `ProductionPollutionPercent` execute through city crime/pollution calculations and effective-building-aware city effects; `NuclearPlant` production blocking and cleanup under `EliminateNukes` are covered; flat building `PollutionAmount` is source-classified as absent from released data.

Task 110 update:
- Building protection fields are closed: `ProtectFromNukes`, `ProtectFromBioAgents`, `ProtectFromNanoVirus`, `PreventConversion`, `PreventSlavery`, and `ForceField` execute for direct buildings and wonder-granted effective buildings through the shared city-effect path.

Task 111 update:
- Building military fields are closed: `DefendersPercent`, `CityWalls`, `OffenseBonusLand`, `OffenseBonusWater`, and `OffenseBonusAir` execute through combat modifiers for direct and effective buildings; `Airport` is retained as a source-backed city marker while modeled refuel behavior already follows retail city/airbase refuel semantics.

Task 112 update:
- Effective-building brokerage closure: grouped effective-building effects preserve each building's `Brokerage` + `CommercePercent` relationship so `IncreaseBrokerages` amplifies wonder-granted brokerages exactly like retail's `buildingutil_GetCommercePercent(GetEffectiveBuildings(), ...)` path.
