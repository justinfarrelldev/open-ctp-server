# Task 120 source evidence: terrain transform advance gates and death eligibility

## Retail source checks

- `terrain.cdb` declares `CanDie`, `AddAdvance`, `RemoveAdvance`, `TransformAdd`, `TransformRemove`, and `DeadFood`/`DeadShield`/`DeadGold` terrain modifier fields.
- Released `terrain.txt` uses `AddAdvance`/`RemoveAdvance` next to transform cost/time blocks; e.g. `TERRAIN_PLAINS` gates both transform directions on `ADVANCE_TOOLMAKING` with add time/materials 1/200 and remove time/materials 1/50.
- Retail terrainutil gates terraform by both target add advance and source remove advance:
  - `terrainutil.cpp:380-381` checks `terrToRec->GetAddAdvanceIndex()` and `terrFromRec->GetRemoveAdvanceIndex()` for class terraform tile improvements.
  - `terrainutil.cpp:538-542` rejects direct terrain transforms without source remove and target add advances.
- Retail terrain death metadata is parsed in `TerrainRecord` (`TerrainRecord.h:193`, `TerrainRecord.cpp:209-210`). Ozone/death paths call `Cell::GetCanDie()` before killing (`WrldPoll.cpp:737`, `CityEvent.cpp:648`).

## Server coverage

- Retail terrain import now preserves `CanDie`, `AddAdvance`, `RemoveAdvance`, optional dead-yield values, and transform timing/materials.
- Runtime terrain projection exposes `CanDie`, dead-yield metadata, `TransformAddAdvanceID`, and `TransformRemoveAdvanceID`.
- Terraform legality now requires the actor nation to own the source terrain remove advance and target terrain add advance before queueing unit/public-works transforms.
- Ozone disaster terrain conversion only targets runtime terrains marked `CanDie`; legacy synthetic rules without terrain metadata keep land fallback behavior.

## Validation

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-120-red.txt`
- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-120-green-target.txt`
- Rules GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-120-rules-green.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-120-green.txt`
- Freeze RED/GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-120-freeze-red.txt`, `.sisyphus/evidence/ctp2-retail-progress-task-120-freeze-green.txt`
- Fullgame GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-120-fullgame-green.txt`
