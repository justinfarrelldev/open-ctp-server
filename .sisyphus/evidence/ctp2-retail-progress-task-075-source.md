# Task 075 source check: `HasPopAndCanBuild` units are not normal city production

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/outcom/C3UnitDB.cpp:122-124` returns `FALSE` from unit-buildability checks when the requested unit record has `GetHasPopAndCanBuild()`.
- `ctp2/ctp2_source_code/ctp2_code/ui/interface/EditQueue.cpp:666` skips unit records with `GetHasPopAndCanBuild()` in the queue UI.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/unitutil.cpp:94-112` uses `HasPopAndCanBuild` unit records to identify land/sea city pseudo-unit types, not trainable military/civilian units.

Implementation notes:

- `queueProduction` now rejects unit production items whose runtime unit definition carries `EffectHasPopAndCanBuild`.
- The existing population-removal path for settler-style units with `BuildingRemovesAPop` remains unchanged.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-075-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-075-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-075-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-075-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-075-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-075-diff-check.txt`
