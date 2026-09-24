# CTP2 Full Retail Task 139: public unit target markers and legal actions

## Implementation
- Visible enemy unit projections now expose only safe targetability markers (`can_be_sued`, `can_be_expelled`) rather than full enemy effects.
- Expel action projection now requires a projected `can_be_expelled` marker, avoiding illegal action leaks after engine-side marker enforcement.
- Sue unit action projection now targets visible adjacent/same-tile units with a projected `can_be_sued` marker and uses the `target_unit_id` payload path.

## Tests
- `go test ./internal/gameplay/visibility -run TestRetailUnitTargetActionsUsePublicTargetMarkers -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Richer visible unit identity/class projections are still intentionally omitted until a full disclosure audit is completed.
