# Task 10 CTP2 source inventory: movement, combat, armies, transports, special action integration

Source root inspected before and after implementation: `/home/justi/Projects/open-ctp-server/ctp2/ctp2_source_code`.

## Armies and movement domains
- `ctp2_code/gs/gameobj/ArmyData.cpp:494` (`ArmyData::GetMovementType`) intersects movement types across army members.
- `ctp2_code/gs/gameobj/ArmyData.cpp:506` (`ArmyData::GetCargoMovementType`) calculates cargo movement separately from the transport army movement.
- Server alignment: army movement still intersects army-member domains, but loaded transports now move using the transport's own domain; cargo domains are checked for embark/disembark instead of blocking transport traversal.

## Cargo / unloading
- `ctp2_code/gs/gameobj/UnitData.cpp:839` (`UnitData::CanThisCargoUnloadAt`) checks movement points, destination stacking, enemy occupation/beach-assault eligibility, ZOC, city capture restrictions, and destination terrain entry.
- Server alignment: disembark uses effective terrain, stack limits, access checks, movement cost, and rejects enemy-occupied destinations in this bounded no-beach-assault slice.

## Combat
- `ctp2_code/gs/gameobj/ArmyData.cpp:1105` (`ArmyData::CanFight`) requires at least one combat-capable attacker and target-domain/city compatibility before resolving combat.
- Server alignment: combat now rejects non-combat armies and incompatible domain attacks before RNG, while preserving deterministic aggregate combat results.

## Special actions and embassies
- `ctp2_code/gs/gameobj/UnitData.cpp:5459` / `5498` require the relevant special capability, adjacency, and movement points for embassy/steal actions.
- `ctp2_code/gs/gameobj/ArmyData.cpp:2536` and `UnitData.cpp:4129` establish embassies through durable player relationship state and reject own-city targets.
- Server alignment: unit special actions already spend movement and require adjacency/capability; embassy now also creates durable contact in the canonical diplomacy relation table.
