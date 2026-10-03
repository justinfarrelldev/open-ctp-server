# Task 121 source evidence: tile-improvement metadata, detection, border radius

## Retail source checks

- `TerrainImprovementRecord` parses gameplay-relevant metadata: `Class`, `Level`, `ConstructionTiles`, `CanSee`, `IntBorderRadius`, `SquaredBorderRadius`, effect flags (`Airport`, `Radar`, `ListeningPost`, `Endgame`), `Freight`, terrain effects, and construction costs (`TerrainImprovementRecord.h`, `TerrainImprovementRecord.cpp`).
- Completed installations call `terrainutil_DoVision()` and then, when both border radii are present, `GenerateBorders(m_point, m_owner, intRad, sqRad)` (`TerrImproveData.cpp:113-118`).
- `GenerateBorders` claims the center tile and expands through neutral/current-owner tiles within squared radius without crossing another owner (`CityInfluenceIterator.cpp:289-344`).
- Installation detection consults `GetCanSee()` bit intersections for installed detectors and unit visibility/detection (`installationdata.cpp:150-160`, `terrainutil.cpp:349`).
- Released `tileimp.txt` uses these fields on forts/listening posts/roads/railroads, e.g. fortifications have `IntBorderRadius 2`, `SquaredBorderRadius 5`, construction tiles, defense/heal/vision, and `CanUpgrade`; listening posts have `CanSee:Standard`; ocean detectors include `CanSee:Underwater`; roads/railroads have `Freight` values. Retail `World::CalcTerrainFreightCost` is a constant stub, so freight is preserved/projection metadata rather than an active path multiplier in the released source.

## Server coverage

- Retail import/runtime projection now preserves level, construction tiles, can-see bits, border radii, and freight metadata for terrain improvements.
- Built overlays now carry can-see, freight, and border metadata.
- Completing a border-radius terrain improvement claims neutral/current-owner tiles within the source-backed radius shape and does not overwrite enemy-owned tiles.
- Completed detector overlays with `CanSee:Underwater` can detect submarines for projection visibility. `CanSee:Stealth` is supported generically although released `tileimp.txt` does not contain a stealth detector.
- `CanUpgrade` remains a source-backed marker: released CTP2 source exposes parser/accessor data but no upgrade command call site was found for tile improvements; server preserves the marker without inventing a hidden upgrade mechanic.

## Validation

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-121-red.txt`
- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-121-green-target.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-121-green.txt`
- Freeze RED/GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-121-freeze-red.txt`, `.sisyphus/evidence/ctp2-retail-progress-task-121-freeze-green.txt`
- Fullgame GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-121-fullgame-green.txt`
