# CTP2 Full Retail Task 188: professional units are support-free

## Source evidence
- `Readiness.cpp::GetSupportCost` returns `0` when `u.GetIsProfessional()` is true.
- `Readiness.cpp::SupportUnit` and `KillUnitsOverBudget` both use `GetSupportCost`, so professional units do not contribute to normal readiness support cost.
- `UnitData` persists the professional flag as `k_UDF_IS_PROFESSIONAL`.

## Implementation
- Added canonical `Unit.IsProfessional` with safe public projection for visible units.
- Readiness support candidate collection now skips professional units before computing shield-hunger support cost.
- Legacy units without the flag retain existing non-professional behavior.

## Tests
- `go test ./internal/gameplay/engine -run TestReadinessSupportIgnoresProfessionalUnits -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
