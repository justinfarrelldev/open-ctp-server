# CTP2 full-retail parity task 14 unit fuel evidence

Scope: project retail unit fuel and enforce a bounded fuel-spend rule during movement.

Original source/data checked:
- `ctp2_code/gs/newdb/unit.cdb`/`UnitRecord` define `MaxFuel` and fuel-related flags.
- `ctp2_code/gs/gameobj/UnitData.cpp` initializes unit fuel from `GetMaxFuel()` and deducts fuel during movement for units with fuel/crash behavior.
- Retail data example verified: `UNIT_FIGHTER` has `MovementType Air`, `NoFuelThenCrash`, and `MaxFuel 2400`.

Implemented server-compatible slice:
- Runtime unit definitions and starter projections carry `max_fuel`.
- Retail projection maps `MaxFuel`; synthetic compile accepts optional `MaxFuel` and air/space domains.
- Engine unit state stores `fuel`/`max_fuel`; produced and starting units initialize fuel to max.
- Movement now rejects fueled units without enough fuel for the path and deducts fuel by movement cost for successful moves.
- Engine validation accepts retail air/space domains and validates fuel invariants.

Validation commands:
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run UnitMovementSpendsFuelAndRejectsEmptyFuel -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/game -count=1`
