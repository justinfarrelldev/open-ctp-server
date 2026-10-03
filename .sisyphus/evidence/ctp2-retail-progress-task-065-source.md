# Task 065 source check: bombard building destruction removes illegal queue items

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:4281-4283` calls `c.DestroyRandomBuilding()` when city bombard building destruction succeeds.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:4746-4779` implements `CityData::DestroyRandomBuilding`; after clearing the selected non-capitol improvement, it calls `m_build_queue.RemoveIllegalItems(TRUE)`.

Implementation notes:

- After city bombard destroys a building, the server now revalidates the city production queue with the runtime production unlock rules and removes queued items that became illegal (for example, items requiring the destroyed building).
- Still-legal queued items are preserved in original order.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-065-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-065-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-065-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-065-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-065-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-065-diff-check.txt`
