# CTP2 Full Retail Task 164: unit support fields

## Source evidence
- Retail `Units.txt` defines `ShieldHunger` and flags such as `NeedsNoSupport` / `IsSpecialForces`.
- `Readiness.cpp` uses `GetShieldHunger()`, `GetNeedsNoSupport()`, and `GetIsSpecialForces()` when computing military support cost.

## Implementation
- Runtime unit definitions now preserve `ShieldHunger`, `NeedsNoSupport`, and `IsSpecialForces` from retail import and synthetic rules.
- Produced canonical units copy those support fields.
- Owner-safe unit projections expose the fields only for the owning player.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
