# CTP2 Full Retail Task 166: readiness over-budget unit disband

## Source evidence
- `Readiness.cpp::KillUnitsOverBudget` removes units when readiness support cost exceeds available military production plus franchise production.
- Units are considered by support cost; unsupported units are skipped and higher-cost supported units are removed first.

## Implementation
- Readiness support tax now disbands supported units deterministically when support cost exceeds available production budget.
- Highest support-cost units are removed first, with ID tie-breaking for deterministic server outcomes.
- Support cost is recalculated after removals before charging city production.

## Tests
- `go test ./internal/gameplay/engine -run TestReadinessSupportDisbandsHighestCostUnitsOverProductionBudget -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail professional-unit ordering and user-facing no-support notifications remain future parity work.
