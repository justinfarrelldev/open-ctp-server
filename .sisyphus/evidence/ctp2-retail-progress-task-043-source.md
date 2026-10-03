# Task 043: retail EnableParkRangers global production unlock

## Retail source/data checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:4212-4217`
  - City unit production rejects records with `CreateParks` unless `wonderutil_GetParkRangersEnabled(g_theWonderTracker->GetBuiltWonders())` is true.
  - The check uses the global wonder tracker built-wonder mask, not only the prospective builder's own wonders.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:10238-10240`
  - The player production path applies the same global `wonderutil_GetParkRangersEnabled(g_theWonderTracker->GetBuiltWonders())` gate.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/wonderutil.cpp:110-112`
  - `wonderutil_GetParkRangersEnabled` reads active `EnableParkRangers` wonder flags.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Wonder.txt:314-322`
  - `WONDER_THE_EDEN_PROJECT` has `EnableParkRangers`.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Units.txt:1417`
  - Retail park-creation units use `CreateParks`.

## Implementation notes

- Added an engine helper for global active wonder markers.
- Park-ranger/`CreateParks` unit production now matches retail: it remains locked until any non-destroyed, non-obsolete `EnableParkRangers` wonder exists, regardless of owner.
- The no-wonder rejection remains intact.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-043-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/engine -run TestRetail.*ParkRanger -count=1`
