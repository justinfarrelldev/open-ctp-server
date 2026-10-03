# Task 045: retail EliminateNukes blocks nuclear-plant buildings

## Retail source/data checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:4294-4299`
  - City building production rejects a building record with `NuclearPlant` when `wonderutil_GetNukesEliminated(g_theWonderTracker->GetBuiltWonders())` is true.
  - The source checks the global built-wonder mask, not just the city owner's wonders.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/wonderutil.cpp`
  - `GetNukesEliminated` is the wonder flag backing retail `EliminateNukes`.
- Retail runtime import already maps building `NuclearPlant` to `EffectNuclearPlant` and wonder `EliminateNukes` to `EffectEliminateNukes`.

## Implementation notes

- Building queue validation now rejects `ProductionBuilding` items with `EffectNuclearPlant` once any active, non-obsolete `EffectEliminateNukes` wonder exists.
- Nuclear plant buildings remain queueable before that global wonder condition appears.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-045-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/engine -run TestRetail.*NuclearPlant -count=1`
