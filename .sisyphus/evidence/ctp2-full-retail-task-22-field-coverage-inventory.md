# CTP2 full-retail parity task 22 field coverage inventory

Scope: inventory every unique released-game gameplay/catalog field currently parsed from the retail CTP2 data files, and record high-level runtime/execution coverage status. This is an inventory only; it is not a full-parity claim.

Source data root:
`ctp2/ctp2_released_game/ctp2_data/default/gamedata`

Regression guard:
- `internal/gameplay/rules/retail_field_inventory_test.go`
- `TestRetailFieldInventorySnapshot` verifies the released catalog field counts and sentinel behavior-affecting fields.

## Summary counts

| Catalog | Records | Unique fields, including nested fields | Current coverage summary |
| --- | ---: | ---: | --- |
| Units | 79 | 194 | Core cost/prereq/obsolete/combat/move/fuel/ranged/bombard plus `SpaceLaunch`, `NuclearAttack`, `ThrowParty`, `Advertise`, `CanBeGifted`, `CanPirate`, `DeathEffectsHappy`, `InvisibleForGlobeSat`, and beach-assault unload movement behavior are partially projected/executed. Many special actions, cargo/vision classes, stealth, active defense, upgrades, support, and sound/UI fields remain pending. |
| Advances | 107 | 19 | Cost/prerequisites, research unlocks, production obsolescence, plant-nuke unlocks, pollution modifiers, capitalization/infrastructure unlocks, terrain-improvement transport/terraform gates, and advance-triggered feats are executed; Age/Branch/Icon/GLHidden/advice are presentation metadata. |
| Feats | 14 | 13 | Advance-triggered feat effects for production, science, happiness, boat movement, hit points, and city-wall reduction are imported as timed runtime effects; building thresholds, sailed-around-world, conquered-by-force, and founder city-recapture triggers now execute as timed nation effects with retail one-time achievement gating. |
| Wonders | 30 | 39 | Cost/prereq/obsolete metadata plus selected effects including `IncreaseBoatMovement`, `IncreaseHp`, `IncreaseScientists`, `IncreaseBrokerages`, `IncreaseRegard`, treasury-level gold (`BonusGold`, route gold, television gold), `EnableParkRangers`, `EliminateNukes`, and `FreeSlaves` are projected/executed; Gaia and remaining global effects require more execution coverage. |
| Buildings | 52 | 42 | Cost/upkeep/prereq/obsolete, production/gold percent effects, `DoubleTelevangelists` conversion gating, and `NuclearPlant` production blocking under `EliminateNukes` are projected/executed. Science, happiness, crime, population, pollution, protection, combat, trade, and growth effects remain pending. |
| Terrain improvements | 40 | 36 | Class, construction cost/time, movement/yield effects, CantBuildOn, and Excludes are partially supported. Vision/radar/heal/terraform/endgame/listening-post behavior remains pending. |
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
