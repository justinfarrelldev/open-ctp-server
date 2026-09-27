# CTP2 Full Retail Tasks 352-358: advance flags, pollution modifiers, and improvement advance gates

## Source notes
- Retail `Advance.txt` has gameplay flags including `Capitalization`, `Infrastructure`, `Transform`, `Roads`, `Railroads`, `MagLev`, `Tunnels`, `DeepOcean`, plus pollution modifiers.
- Retail `AdvanceRecord` stores `PollutionSizeModifier` and `PollutionProductionModifier` as advance fields.
- Retail terrain improvement `TerrainEffect` blocks can name `EnableAdvance` (for example farms require `ADVANCE_AGRICULTURE`).

## Implementation
- Runtime advances now preserve the gameplay flags above and signed pollution modifier percentages.
- Retail city pollution applies known advance pollution-size modifiers to population pollution and production modifiers to industrial pollution.
- Runtime terrain improvements now preserve `required_advance_id` from retail `TerrainEffect.EnableAdvance`.
- Canonical terrain-improvement execution rejects advance-gated improvements until the actor nation knows the required advance.
- Safe terrain-action projection suppresses advance-gated build-improvement actions until the viewer's owner research projection includes the required advance, so AI consumes only legal projected actions.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestRetailTerrainImprovementRequiresEnabledAdvance -count=1`
- `go test ./internal/gameplay/engine -run TestRetailCityPollutionAppliesAdvancePollutionModifiers -count=1`
- `go test ./internal/gameplay/visibility -run TestTerrainActionsSuppressAdvanceGatedImprovements -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
- `go test ./internal/gameplay/ai -count=1`
