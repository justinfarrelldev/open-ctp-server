# Task 129 source evidence: modeled retail constants closure

## Retail source checks

- `Const.txt` defines modeled gameplay constants used by city growth, special actions, indulgence/conversion, emancipation unhappiness, nukes, piracy waste, veteran combat coefficient, bombard probabilities, special-action move cost, and caravan/trader transport coefficient.
- Retail call sites include special action execution in `UnitData.cpp`/`unitevent.cpp`, nuke pollution/population effects in combat/unit event paths, trade route/freight calculations in `TradeRouteData.cpp`/`tradeutil.cpp`, piracy in route/unit paths, bombard outcomes in combat/city event paths, and city growth/happiness in city/happy code.

## Server coverage/classification

- Runtime imports the modeled `Const.txt` constants and tests their released values.
- Engine execution covers modeled constants: city growth coefficient, cleric/televangelist conversion factors, indulgence gold/happiness, emancipation unhappiness, nuke population/kill-tile/pollution, piracy waste, veteran coefficient, bombard building/population probabilities, `SPECIAL_ACTION_MOVE_COST`, and `CARAVAN_COEF` trade transport cost.
- Remaining `Const.txt` values not projected into runtime are source-classified as UI/audio/tuning/worldgen/AI or mechanics outside this server's canonical modeled state; no behavior-bearing imported runtime constant remains parse-only.

## Validation

- GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-129-green.txt`
