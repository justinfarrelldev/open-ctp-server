# Task 052: retail beach-assault unload movement

## Retail source checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:6438-6451`
  - After unloading cargo, retail queues `GEV_SetUnloadMovementUnit` only for debarked units that are not `k_UDF_BEACH_ASSAULT_LEGAL` and are not air units.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:6846-6855`
  - `SetUnloadMovementPoints()` sets movement to zero for non-beach-assault, non-air units after unload.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:4982-4994` and `5892-5910`
  - Leaving a transport records whether beach assault is legal. `CanBeachAssaultRightNow()` is true for paratroopers carried by paratrooper transports or beach-assault units carried by sea/shallow-water transports.

## Implementation notes

- Disembarking still requires enough movement to enter the destination tile.
- Non-air cargo without a currently legal beach-assault/paratrooper unload now has movement set to zero after disembarking.
- Air cargo, paratroopers from paratrooper transports, and `CanBeachAssault` units from sea transports retain their post-entry remaining movement.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-052-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/engine -run 'TestRetailDisembarkZeroesMovementForNonBeachAssaultCargo|TestRetailBeachAssaultCargoKeepsPostMoveMovementAfterSeaDisembark|TestTask8CargoFortifyDisbandAndArmyVectors' -count=1`
