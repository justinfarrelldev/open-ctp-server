# Task 096 source notes — combat stat/modifier coverage verified

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/UnitRecord.cpp:1949-2052` parses `BombRounds`, `WoodenShipBonus`, and `BombardRange`; `UnitRecord.h:1355-1475` exposes `Firepower`, `Armor`, `BombRounds`, `WoodenShipBonus`, and `BombardRange`.
- `ctp2_source_code/ctp2_code/gs/gameobj/CTP2Combat.cpp:864-865` constructs combat units from retail `Attack`, `Defense`, `Firepower`, `Armor`, `ZBRangeAttack`, and HP values.
- `ctp2_source_code/ctp2_code/gs/gameobj/CTP2Combat.cpp:1052,1155` computes damage from attacker strength against defender armor.
- `ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:1267,1376-1379,5108-5119` uses bombard range, firepower/armor, bombard rounds, and firepower for bombard damage.
- `ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:1358` applies `AttackWoodenShipBonus`/`WoodenShipBonus` against wooden-ship defenders.
- `ctp2_source_code/ctp2_code/gs/gameobj/unitutil.cpp:156-194`, `UnitData.cpp:3539-3606`, `CityData.cpp:3038-3040`, and `buildingutil.cpp:267-278` cover city-wall detection/defense and ignore/nullify wall behavior.
- `ctp2_source_code/ctp2_code/gs/newdb/FeatRecord.cpp:525-527` and `FeatTracker.cpp:188,333` parse/apply `EffectReduceCityWalls`.

Server parity already in place:

- Retail unit import/runtime projection preserves `Attack`, `Defense`, `ZBRangeAttack`, `BombRounds`, `BombardRange`, `Firepower`, `Armor`, `MaxHP`, `MountedBonus`, `BonusAgainstMounted`, `AttackWoodenShipBonus`, `WoodenShipBonus`, `Defend`, `CanAttack`, `CanBombard`, `CanCounterBombard`, `CityWalls`, `ReduceCityWalls`, and `IgnoreCityWalls` into engine stats/effects.
- Engine combat/bombard resolution consumes those stats/effects in grouped combat, bombard, counter-bombard, city-wall defense, wooden-ship attack bonus, and mounted-defense modifier paths.

Validation:

- `go test ./internal/gameplay/engine -run 'TestRetailFirepowerScalesCombatDamage|TestRetailFirepowerScalesBombardDamage|TestRetailCanAttackEffectsRestrictCombatTargets|TestRetailBombardUsesCanBombardDomains|TestRetailBombardDomainMasksSelectEligibleTargets|TestRetailCanCounterBombardReturnsFireAfterBombardment|TestRetailCounterBombardConsidersAllDefendersOnTargetTile|TestRetailCounterBombardHonorsDefenderBombardRange|TestRetailIgnoreCityWallsRemovesOnlyWallDefense|TestRetailMountedDefenseBonusModifiesDefenderAgainstMountedAttackers|TestRetailWoodenShipAttackBonusAppliesAgainstWoodenShips|TestRetailCityBombardWallDefenseUsesActualBuildingsOnly|TestAdvanceFeatReduceCityWallsReducesDefenderModifier' -count=1`
- `go test ./internal/gameplay/rules -run 'TestProjectRetailRuntimeMapsSelectedRetailValues|TestLoadRetailRulesFromDirectoryPreservesTypedRetailValues' -count=1`
