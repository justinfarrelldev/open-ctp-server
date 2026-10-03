# Task 125 source evidence: pollution trigger map-size selection

## Retail source checks

- Released `pollution.txt` contains four trigger lists keyed by map-size comments:
  - `POLLUTION_TRIGGER_LIST_SMALL_MAP` for `MAP_SIZE 24 48` (1152 tiles)
  - `POLLUTION_TRIGGER_LIST_MEDIUM_MAP` for `MAP_SIZE 48 96` (4608 tiles)
  - `POLLUTION_TRIGGER_LIST_LARGE_MAP` for `MAP_SIZE 64 128` (8192 tiles)
  - `POLLUTION_TRIGGER_LIST_GIGANTIC_MAP` for `MAP_SIZE 70 140` (9800 tiles)
- Retail setup captures map size before world construction (`gameinit.cpp` map-size paths), and pollution trigger data is indexed by that effective world size.

## Server coverage

- Runtime pollution now chooses the trigger list from the canonical world's tile count instead of always selecting the small-map list.
- Thresholds follow released data tile counts; unsupported/missing larger lists fall back toward smaller lists and retain a `HUGE` alias for existing server fixtures.
- Existing pollution phase/history/disaster behavior remains unchanged after trigger selection.

## Validation

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-125-red.txt`
- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-125-green-target.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-125-green.txt`
- Freeze RED/GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-125-freeze-red.txt`, `.sisyphus/evidence/ctp2-retail-progress-task-125-freeze-green.txt`
- Fullgame GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-125-fullgame-green.txt`
