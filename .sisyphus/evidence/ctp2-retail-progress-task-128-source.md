# Task 128 source evidence: victory, endgame, space, Gaia/final turn

## Retail source checks

- Retail victory/endgame logic spans final-turn/calendar (`TurnCnt.cpp`, `newturncount.cpp`), bloodlust/conquest player lifecycle, diplomatic/world-union paths, score/ranking terminal state, Gaia/Solaris endgame objects and countdown (`gaiacontroller.cpp`, `EndGameObjects.txt`, `FeatTracker.cpp`, `PlayerEvent.cpp`).
- Space-launch interactions use `SpaceLaunch` unit effects and endgame/wonder/object state (`UnitData.cpp`, `unitevent.cpp`, `bldque.cpp`).
- Wonder capture/destruction/obsolescence affects endgame enablement, global radar, building-everywhere effects, and Gaia controller status (`wonderutil.cpp`, `CityEvent.cpp`).

## Server coverage

- Runtime imports final-turn/calendar, victory evaluator IDs, endgame object definitions, and wonder/endgame markers.
- Engine coverage includes conquest/bloodlust terminal state, diplomatic/world-union victory, score/final-turn evaluation, Gaia controller countdown prerequisites and interruption/restart conditions, space-launch unit effects, endgame object loss, dead-player lifecycle, rankings/tie-breaking, and terminal event sequencing.
- Visibility/fullgame coverage validates projections and restart equality for the victory/endgame scenarios.

## Validation

- GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-128-green.txt`
