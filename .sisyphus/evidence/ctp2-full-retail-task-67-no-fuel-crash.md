# CTP2 Full Retail Task 67: no-fuel crash air units

## Source references
- Retail `Units.txt` marks aircraft/missiles such as `UNIT_BOMBER` with `NoFuelThenCrash`.
- `unit.cdb` defines `NoFuelThenCrash` as a unit behavior flag.

## Implementation
- Retail unit projection maps `NoFuelThenCrash` to a runtime `no_fuel_then_crash` effect.
- Turn advancement now refuels eligible units first, then removes stranded fueled units whose fuel remains zero and which carry the crash effect.

## Tests
- `go test ./internal/gameplay/engine -run TestNoFuelThenCrashRemovesStrandedAirUnitsOnAdvanceRound -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules -count=1`

## Remaining gaps
- Full retail fuel behavior still needs detailed crash messaging, carrier/base nuances, and AI safety planning around fuel ranges.
