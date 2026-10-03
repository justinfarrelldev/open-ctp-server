# Task 126 source evidence: pollution accumulation, reducers, flood/ozone outcomes

## Retail source checks

- Retail city pollution comes from population and production with `PollutionModel` thresholds/ratios and is modified by governments, advances, buildings, and wonders (`CityData.cpp`, `Pollution.cpp`, `GovernmentRecord`, `BuildingRecord`, `WonderRecord`).
- Retail disasters use pollution phase/history and trigger-list thresholds; flood converts a land tile to shallow water and removes occupants/improvements, while ozone kills eligible land into dead terrain and cuts improvements/goods (`WrldPoll.cpp:737`, flood/ozone paths; `Cell::Kill`; `World::ClearGoods`; `World::CutImprovements`).
- Retail wonders/advances include pollution reduction and pollution-unhappiness modifiers; nuclear effects add/transform pollution in attack paths.

## Server coverage

- City pollution accumulation covers population/production model thresholds, government coefficient, building/wonder/advance pollution modifiers, nuclear plant/EliminateNukes cleanup, tile pollution totals, and pollution-unhappiness suppression/reduction paths.
- Global pollution phase/history/trend and disaster event selection are wired to runtime trigger lists and tile-count map-size selection.
- Flood disaster converts the selected polluted land tile to shallow water, clears owner/pollution, removes cities/units/overlays on that tile, and emits pollution state metadata.
- Ozone disaster selects only runtime `CanDie` terrain, converts to dead terrain, clears tile pollution, and removes overlays.

## Validation

- GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-126-green.txt`
