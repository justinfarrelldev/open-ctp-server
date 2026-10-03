# Task 087 source notes — building icon/description metadata classification

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/BuildingRecord.cpp:492-500` parses `DefaultIcon` as an Icon DB record and `Description` as a string ID.
- `ctp2_source_code/ctp2_code/gs/newdb/BuildingRecord.h:139-141` exposes `GetDefaultIconIndex()`, `GetDefaultIcon()`, and `GetDescription()` accessors.
- `ctp2_source_code/ctp2_code/gs/newdb/BuildingRecord.cpp:969-972` resolves `GetDefaultIcon()` through `g_theIconDB`.
- `ctp2_source_code/ctp2_code/ui/interface/EditQueue.cpp:1528-1538` uses building `DefaultIcon` for production queue item display, alongside unit/wonder icons.
- `ctp2_source_code/ctp2_code/ui/interface/scenarioeditor.cpp:734-741` uses building icon records to set scenario-editor switch images.
- Building description/name references are routed through SLIC/UI contexts such as `EditQueue.cpp:1533-1534`, `citywindow.cpp`, and `greatlibrary.cpp`; no gameplay state mutation or legality branch is tied to building `Description`.

Classification:

- Building `DefaultIcon` is presentation/setup metadata for server gameplay.
- Building `Description` is text/UI/SLIC metadata for server gameplay.

Validation:

- `go test ./internal/gameplay/rules -run TestRetailFieldInventorySnapshot -count=1`
