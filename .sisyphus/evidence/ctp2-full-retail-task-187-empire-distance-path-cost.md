# CTP2 Full Retail Task 187: empire-distance path costs

## Source evidence
- `WorldDistance.cpp::FindCityDistances` computes city distance from the capital with a priority queue over tile movement costs.
- `WorldDistance.cpp::CDMove` stores the accumulated search count into `Happy::SetCostToCapitol` for each owned city.
- `Happy.cpp::CalcDistanceFromCapitol` uses that stored movement cost, applies difficulty adjustment, clamps to government max distance, subtracts minimum distance, and scales by `EmpireDistanceScale`.

## Implementation
- Empire-distance happiness now uses terrain movement-cost pathing from the capital city to the target city when pinned runtime terrain movement data is available.
- The path cost includes the capital tile and destination tile movement costs, matching the retail accumulated search-count model.
- If runtime movement data is unavailable or incomplete, the previous deterministic Manhattan fallback remains active.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailEmpireDistanceUsesTerrainPathCosts -count=1`
- `go test ./internal/gameplay/engine -count=1`
