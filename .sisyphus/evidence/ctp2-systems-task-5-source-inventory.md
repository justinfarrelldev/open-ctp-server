# Task 5 CTP2 source inventory: unit capabilities and movement

Source root inspected: `/home/justi/Projects/open-ctp-server/ctp2/ctp2_source_code` (Activision CTP2 source release). Task 5 stays synthetic; no retail data or literals were copied.

## Citations and implementation consequences

| Source path | Lines / symbol | Consequence in server |
|---|---|---|
| `ctp2_code/gs/newdb/UnitRecord.h` | `m_MovementType` and `GetMovementType*` around lines 747 and 877-884 | Runtime unit definitions preserve a bounded `Domain` / movement capability instead of treating all units as generic land units. |
| `ctp2_code/gs/newdb/UnitRecord.h` | `CargoData::GetMaxCargo` line 683 and `GetCanCarry*` accessors around lines 889-904 | Runtime unit definitions include `TransportCapacity` and `CargoDomains`, with invalid cargo-domain combinations rejected. |
| `ctp2_code/gs/newdb/UnitRecord.h` | `GetShieldCost`, `GetMaxHP`, `GetAttack`, `GetDefense`, `GetArmor`, `GetVisionRange`, `GetMaxMovePoints` around lines 1349-1378 | Runtime unit definitions preserve production cost, health, attack/defense, armor, visibility range, and movement budget as deterministic synthetic fields. |
| `ctp2_code/gs/newdb/UnitRecord.h` | `GetEnableAdvanceIndex` / `GetObsoleteAdvanceIndex` around lines 905-908 | Runtime unit definitions preserve prerequisite and obsolete-by IDs for Task 6 unlock semantics, without requiring retail advance names. |
| `ctp2_code/gs/gameobj/UnitData.cpp` | Constructor line 276 and `UnitData::ResetMovement` around line 592 | Start-game and round refresh derive movement from the compiled unit definition / pinned runtime, matching source behavior where movement comes from `UnitRecord::GetMaxMovePoints`. |
| `ctp2_code/gs/gameobj/Unit.cpp` | Cargo movement-type checks around line 2543 | Server cargo compatibility remains domain-based and deterministic; no size-specific retail table was copied. |
| `ctp2_code/gs/world/Cell.cpp` | `Cell::CanEnter` line 257 | Server terrain entry checks continue to intersect unit movement domains with terrain movement domains. |
| `ctp2_code/gs/world/WrlEnv.cpp` | `World::GetMoveCost` line 437 | Terrain movement cost remains a tile/runtime terrain property, not a hidden unit-side shortcut. |

## Boundaries

- Full retail unit catalog parity is not claimed.
- Unit special effects are a bounded synthetic vector (`settler`, embassy/steal/capture/terraform hooks) used by existing server commands.
- Movement domains are intentionally reduced to the server-supported set: `land`, `sea`, and `amphibious`; air/space/mountain-specific CTP2 movement remain future matrix work.
- Buildings, wonders, terrain improvements, governments, trade, and pollution still require their own Task 6+ source-audited implementation gates before final completion.
