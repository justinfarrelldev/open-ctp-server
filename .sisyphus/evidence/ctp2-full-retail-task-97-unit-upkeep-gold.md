# CTP2 Full Retail Task 97: unit upkeep gold support

## Source references
- Retail unit `GoldHunger` is per-unit gold support/maintenance.
- Retail unit upkeep should affect the owning nation's economy/research gold, and must survive unit production.

## Implementation
- Canonical units now store `upkeep` and validation rejects negative unit upkeep.
- Runtime unit production copies retail/runtime unit upkeep into produced units.
- Owner projections expose owned-unit upkeep for safe planning.
- Research rate now subtracts upkeep for units owned by the nation before applying nation research percent effects, clamped at zero city-gold contribution.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailResearchRateSubtractsUnitUpkeep|TestProductionCompletionCopiesRuntimeRangedStats' -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Full treasury, support-free-unit rules, bankruptcy/disband, readiness support, and slider allocation remain to be implemented.
