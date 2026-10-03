# Task 082 source classification: unit presentation/index metadata

Retail source evidence:

- `NoIndex`: `ctp2/ctp2_source_code/ctp2_code/ui/netshell/ns_units.cpp:54` stores `GetNoIndex()` into the setup UI unit list, and `ui/netshell/allinonewindow.cpp:1481-1510` uses it to hide unit-exclusion checkboxes. This is setup/presentation, not in-game production legality.
- `GLHidden`: `ctp2/ctp2_source_code/ctp2_code/ui/interface/greatlibrary.cpp:2172-2182` skips Great Library topics with `GetGLHidden()`. This is encyclopedia presentation only.
- `DefaultSprite`: `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:308` creates a `SpriteState` from `GetDefaultSprite()->GetValue()` for actor rendering.
- Unit sound fields: `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Unit.cpp:992-1054` exposes record sound indices; `ArmyData.cpp:2023`, `:6193-6195`, `:6365`, and `:6431` pass those IDs to director/sound paths for move/load/cant-move/unload presentation.
- `DefaultIcon` and the `Sound*`/sprite fields are declared in `newdb/unit.cdb:411-416`, `:465` and stored as icon/sound/sprite database indices.

Classification:

- Unit `NoIndex`, `GLHidden`, `DefaultIcon`, `DefaultSprite`, and `Sound*` are presentation/setup metadata for literal server gameplay. They should remain safely imported/projected when clients need catalog presentation, but they do not alter production legality, combat, movement, economy, or persistence mechanics.

Validation evidence:

- Source grep: `.sisyphus/evidence/ctp2-retail-progress-task-082-grep.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-082-diff-check.txt`
