# CTP2 Full Retail Task 81: slave raid special actions

## Source references
- Retail unit fields include `SlaveRaids` and `SettlerSlaveRaids`.
- Retail building fields include `PreventSlavery` protection percentage.

## Implementation
- Added `slave-raid` and `settler-slave-raid` commands with disclosure-safe success/failure events.
- Capable adjacent owned units spend movement to raid enemy cities.
- City-local `prevent_slavery_percent` is resolved with deterministic special-action RNG and can block the raid.
- Successful raids reduce target city population by one without dropping below population one.
- Owner legal-action projections expose visible adjacent raid targets for capable units.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailSlaveRaidsReducePopulationAndHonorProtection -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Captured-slave unit/state modeling, diplomacy, abolition/prohibit-slaver wonders, and exact retail raid probabilities still need implementation.
