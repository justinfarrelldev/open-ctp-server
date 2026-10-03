# Task 117 source evidence: visibility/endgame wonder effects

## Retail source checks

- Retail parses `GlobalRadar`, `StartGaiaController`, and `BuildingEverywhere` in `WonderRecord` (`ctp2_source_code/ctp2_code/gs/newdb/WonderRecord.cpp:224`, `274-277`, `602-603`, `770-778`).
- Retail global-radar state is stored on the player and recomputed from live built wonders (`gameinit.cpp:2493`, `Player.cpp:7594-7595`, `7692-7693`, `9902-9903`) and used by vision/tile/unit visibility (`Vision.cpp:159`, `213`, `228`, `642`; `WrlUnit.cpp:507-516`).
- Retail Gaia controller helpers identify `StartGaiaController` wonders (`wonderutil.cpp:250-252`, `317`, `389`) and the released server runtime maps the Gaia wonder/building/tower prerequisites into the countdown model.
- Retail `BuildingEverywhere` is parsed as a building-record reference (`WonderRecord.cpp:776-778`, `1033-1035`) and source paths use effective buildings broadly; released data also contains `ActualBuildingEverywhere`, which the server maps to the same effective-building marker because the released source has no separate parser token.

## Server coverage

- `GlobalRadar` reveals all tiles/units through active, non-obsolete wonders and respects `InvisibleForGlobeSat` units.
- `StartGaiaController` drives runtime Gaia prerequisites/countdown/restart/terminal behavior through the Gaia runtime model.
- `BuildingEverywhere` and `ActualBuildingEverywhere` project to effective-building markers; effective building execution across economy, population/happiness, crime/pollution, protection, combat, brokerage, television, and production prerequisites is covered by Tasks 107-112.
- Added `TestRetailActualBuildingEverywhereProjectsEffectiveBuilding` to lock released-data `ActualBuildingEverywhere` projection.

## Validation

- Rules GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-117-rules-green.txt`
- Visibility GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-117-visibility-green.txt`
- Engine GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-117-engine-green.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-117-green.txt`
