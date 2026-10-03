# Task 092 source notes — terrain/tile-improvement presentation metadata classification

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/TerrainRecord.h:187-194,284` exposes terrain `Icon`, `GLHidden`, and `TilesetIndex`; `TerrainRecord.cpp:738-741` resolves `Icon` through the Icon DB.
- `ctp2_source_code/ctp2_code/gs/newdb/TerrainImprovementRecord.h:200,320-323,335-336,353` exposes tile-improvement `TilesetIndex`, `Icon`, `Tooltip`, `Statusbar`, `Column`, and `GLHidden`; `TerrainImprovementRecord.cpp:923-926` resolves `Icon` through the Icon DB.
- `ctp2_source_code/ctp2_code/ui/interface/greatlibrary.cpp:2172-2178` uses `GLHidden` to suppress Great Library topic-list entries.
- `ctp2_source_code/ctp2_code/ui/interface/greatlibrarywindow.cpp:427-451,503-515` uses terrain and terrain-improvement `Icon` data for Great Library presentation.
- `ctp2_source_code/ctp2_code/ui/interface/scenarioeditor.cpp:625-631,883-889` uses terrain/tile-improvement icons for editor switches.
- `ctp2_source_code/ctp2_code/ui/interface/controlpanelwindow.cpp:2770-2798` uses tile-improvement `Column` to lay out control-panel buttons, and `controlpanelwindow.cpp:2850-2878,3850-3866` uses `Icon`, `Statusbar`, `Tooltip`, and terrain `TilesetIndex` for button images/tooltips/status text.
- Tile-improvement `Sound` and `IsCityRuin` have no non-parser gameplay call sites in the released source tree; terrain `HutTilesetIndex*`/tileset fields are tile art selection metadata.

Classification:

- Terrain/tile-improvement `Icon`, `TilesetIndex`, `HutTilesetIndex*`, `Tooltip`, `Statusbar`, `Sound`, `GLHidden`, `IsCityRuin`, and `Column` are presentation/editor/Great-Library/control-panel metadata for server gameplay.

Validation:

- `go test ./internal/gameplay/rules -run TestRetailFieldInventorySnapshot -count=1`
