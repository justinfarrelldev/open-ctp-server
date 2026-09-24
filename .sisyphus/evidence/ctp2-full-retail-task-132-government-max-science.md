# CTP2 Full Retail Task 132: government max science rate

## Implementation
- Retail government `MaxScienceRate` now imports into runtime as `max_science_percent`.
- Government changes attach an active nation effect for the max science cap, including zero-valued caps like anarchy.
- Research calculation now caps the city trade/gold portion used for research while leaving flat city research effects uncapped.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailResearchRateHonorsGovernmentMaxScienceRate -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeIncludesCompleteReleasedCatalog -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`

## Remaining gaps
- Full tax/science/luxury sliders, readiness, wages/rations/workday happiness, and material-pool budgeting remain future government parity work.
