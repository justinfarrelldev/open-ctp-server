# CTP2 Full Retail Tasks 250-253: remaining top-level unit flag marker preservation

## Implementation
- Imported the remaining top-level retail `Units.txt` flags as runtime/canonical marker effects:
  - `NoIndex` -> `no_index`
  - `OnlyBuildOne` -> `only_build_one`
  - `ThrowParty` -> `throw_party`
  - `VictoryEnslavement` -> `victory_enslavement`
- Verified the retail top-level flag audit has no unpreserved marker flags after this batch.
- These markers are preserved for AI/client/runtime parity. Full execution for one-build queue semantics, party diplomacy, and victory enslavement remains separate work.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
