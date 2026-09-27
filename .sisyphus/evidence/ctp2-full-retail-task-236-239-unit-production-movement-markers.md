# CTP2 Full Retail Tasks 236-239: unit movement damage, CantBuild, and trader markers

## Implementation
- Imported retail `LossMoveToDmgNone` as `loss_move_to_dmg_none`.
- Round movement refresh now scales damaged units by current/max health unless they have `loss_move_to_dmg_none`; boat movement bonuses still apply after base refresh.
- Imported retail `CantBuild` as `cant_build`; engine production rejects these runtime unit definitions and safe projections suppress their queue-production actions.
- Imported retail `IsTrader` as `is_trader` for runtime/canonical marker preservation.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailDamagedUnitMovementRefreshScalesByHealthUnlessExempt|TestRuntimeCantBuildUnitCannotBeQueued' -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/visibility -run TestProjectionSuppressesCantBuildRuntimeUnitProduction -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/visibility -count=1`
