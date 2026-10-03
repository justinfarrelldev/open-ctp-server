# Task 105 source notes — movement, ZOC, beach/paratrooper, fuel/refuel, and visibility verified

Retail source evidence:

- `UnitData.cpp:276`, `599-603`, and `5849-5854` use `MaxMovePoints`, damage movement scaling, `LossMoveToDmgNone`, `LossMoveToDmgTwo`, and `VisionRange` from unit records.
- `UnitData.cpp:531-552` deducts movement/fuel and uses `NoFuelThenCrash`; `UnitData.cpp:3402-3404` applies non-space fuel cost proportional to movement points; in the server's 100-point movement scale this is the same cost units spend to move over terrain.
- `CityData.cpp:2765-2797`, `buildingutil.cpp:288-290`, `UnitData.cpp:5662-5687`, and `terrainutil.cpp:620/734` source-verify city/building/tile-improvement airport refuel behavior.
- `UnitData.cpp:5892-5905` implements paratrooper transport and beach-assault checks for preserving movement after cargo exits a transport.
- `UnitData.cpp:232`, `WrlUnit.cpp:162/226/252/614/654/676/698/845`, and `ArmyData.cpp:5708/5728/5804/6723/7760` use `NoZoc`/`IgnoreZOC` style no-ZOC records to skip zone-of-control blocking.
- `unitutil.cpp:32-40`, `terrainutil.cpp:302-343`, `terrainutil.cpp:778-791`, `UnitData.cpp:2806-3082`, `Vision.cpp:159/213/228/642`, `WrlUnit.cpp:505-534`, and `cellunitlist.cpp:303` source-verify unit/tile-improvement vision, global radar, stealth/submarine visibility restrictions, and `InvisibleForGlobeSat` behavior.

Server parity:

- Runtime imports `MovementType`, `MaxMovePoints`, `MaxFuel`, `NoFuelThenCrash`, `LossMoveToDmgNone`, `LossMoveToDmgTwo`, `ImmuneToHostileTerrain`, `NoZoc`/`IgnoreZOC`, `CanBeachAssault`, `Paratrooper`, `ParatrooperTransport`, `VisionRange`, stealth/submarine detector fields, `InvisibleForGlobeSat`, and radar/airport tile-improvement fields.
- Engine movement spends 100-scaled terrain costs, fuel uses the same non-space cost as movement points, air units refuel at friendly cities/airports, no-fuel crash units are removed when stranded, and damage movement scaling honors both loss-move flags.
- Projection exposes only legal embark/disembark/move/attack/special actions from public visibility, hides stale-fog-only targets, and applies global-radar/stealth/submarine/globe-satellite rules without hidden-state shortcuts.
- Space launch/reentry is executed as the source-backed action path; per-turn free-space movement and `SPACE_FUEL_COST` are classified as outside the current server map-movement layer rather than silently approximated.

Validation:

- `go test ./internal/gameplay/engine -run 'TestTask8MoveTerrainZOCAndVisibilityVectors|TestUnitMovementSpendsFuelAndRejectsEmptyFuel|TestAirUnitsRefuelAtFriendlyCitiesOnAdvanceRound|TestNoFuelThenCrashRemovesStrandedAirUnitsOnAdvanceRound|TestRetailTerrainImprovementAirportRefuelsUnitsOnRoundComplete|TestRetailLossMoveToDmgTwoHasMinimumTwoMovement|TestRetailDisembarkZeroesMovementForNonBeachAssaultCargo|TestRetailBeachAssaultCargoKeepsPostMoveMovementAfterSeaDisembark|TestTask10DisembarkRejectsEnemyOccupiedTile|TestSpaceLaunchLeavesMapAndReentersOwnedCity|TestSpaceLaunchKillsUnitWhenDestinationCityIsLost|TestSpaceLaunchRequiresOwnedDestinationCity' -count=1`
- `go test ./internal/gameplay/visibility -run 'TestTerrainImprovementVisionRangeRevealsTiles|TestRetailStealthProjectionRequiresStealthDetection|TestRetailSubmarineProjectionRequiresUnderwaterDetection|TestRetailObsoleteGlobalRadarWonderDoesNotRevealAllTiles|TestRetailGlobalRadarRevealsAllTilesAndUnits|TestRetailInvisibleForGlobeSatUnitHiddenFromGlobalRadarOnly|TestRetailInvisibleForGlobeSatUnitVisibleToNormalSight|TestRetailCargoSizeMismatchSuppressesEmbarkAction' -count=1`
