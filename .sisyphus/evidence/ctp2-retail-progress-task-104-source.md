# Task 104 source notes — `SPECIAL_ACTION_MOVE_COST` execution

Retail source evidence:

- `ConstDB.cpp:565/876/1375` parses `SPECIAL_ACTION_MOVE_COST`; released `Const.txt` sets it to `100.0`.
- `ConstDB.h:548` exposes `SpecialActionMoveCost()`.
- `UnitData.cpp:1230`, `1258`, `2635`, `5414-5559`, and `5808` reject special actions when movement points are below `g_theConstDB->SpecialActionMoveCost()`.
- `UnitData.cpp:4496` and `4503`, plus `armyevent.cpp:905`, deduct exactly `g_theConstDB->SpecialActionMoveCost()` after special-action execution.

Change:

- Imported `SPECIAL_ACTION_MOVE_COST` into `RetailConstants`/`RuntimeConstants` and retail runtime projection.
- All special-action movement spending, including piracy, now uses `runtime.Constants.SpecialActionMoveCost` when provided, with the legacy synthetic fallback of 1 movement point when no runtime constant is pinned.
- Added regression coverage requiring an exact 100 movement points to execute a runtime-pinned special action and rejecting 99.

Validation:

- RED: `go test ./internal/gameplay/engine -run TestRetailSpecialActionMoveCostUsesRuntimeConstant -count=1` failed before the constant/schema existed.
- GREEN: `go test ./internal/gameplay/engine -run 'TestRetailSpecialActionMoveCostUsesRuntimeConstant|TestTask11EmbassySoothsayingAndDisclosureVectors|TestRetailCanPirateProjectsVisibleTouchedForeignTradeRoute|TestRetailCanPirate' -count=1`
- Rules: `go test ./internal/gameplay/rules -run 'TestProjectRetailRuntimeMapsSelectedRetailValues|TestRetailCatalog' -count=1`
- Package: `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/visibility ./internal/gameplay/ai ./internal/gameplay/http -count=1`
- Freeze: `go test ./internal/e2e/freeze -run TestEngineLock -count=1`
