# CTP2 Full Retail Tasks 240-245: additional retail unit marker preservation

## Implementation
- Imported additional retail unit flags as runtime/canonical marker effects:
  - `CanPirate` -> `can_pirate`
  - `Civilian` -> `civilian`
  - `InvisibleForGlobeSat` -> `invisible_for_globe_sat`
  - `NoSlaves` -> `no_slaves`
  - `IsTelevangelist` -> `is_televangelist`
  - `CanBeGifted` -> `can_be_gifted`
- These markers are preserved for AI/client/runtime parity. Full execution for pirate/gift/globe-sat/slave/televangelist-specific behavior remains separate work.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
