# CTP2 Full Retail Tasks 246-249: additional unit marker preservation

## Implementation
- Imported additional retail unit flags as runtime/canonical marker effects:
  - `Advertise` -> `advertise`
  - `GLHidden` -> `gl_hidden`
  - `HasPopAndCanBuild` -> `has_pop_and_can_build`
  - `ImmuneToHostileTerrain` -> `immune_to_hostile_terrain`
- These markers are preserved for AI/client/runtime parity. Full execution for advertising/globe-hidden/pop-build/hostile-terrain semantics remains separate work.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
