# CTP2 Full Retail Task 100: boat movement wonder effect

## Source references
- Retail wonder field `IncreaseBoatMovement` grants additional movement to naval units.
- Retail units have their own movement allowance and should refresh from unit max movement rather than only a scenario starter default.

## Implementation
- Canonical units now store `max_movement`; runtime unit production sets it from the retail/runtime unit movement value.
- Round refresh uses unit `max_movement` when present, with starter movement as a fallback for legacy/synthetic units.
- Sea and amphibious units receive owner/global `increase_boat_movement` wonder bonuses during movement refresh.
- Owner projections expose owned-unit `max_movement` for safe planning.
- Validation rejects negative max movement.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailBoatMovementWonderRefreshesSeaUnitMovement|TestTask17RoundRefreshesMovementFromPinnedRuntime|TestProductionCompletionCopiesRuntimeRangedStats' -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail naval class masks, stacked transport refresh, movement effects from governments/terrain, and veteran/advance modifiers remain to be implemented.
