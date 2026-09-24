# CTP2 Full Retail Task 131: franchise production transfer

## Source evidence
- `ctp2_code/gs/gameobj/CityData.cpp:1219-1244` computes franchise loss as `ceil(net_production * FranchiseEffect)` and calls `g_player[m_franchise_owner]->AddProductionFromFranchise(m_productionLostToFranchise)`.
- `ctp2_code/gs/gameobj/Player.cpp:8501-8514` accumulates `m_productionFromFranchises` for the franchise owner.

## Implementation
- Added `Nation.FranchiseProduction` as the current-turn franchise production accumulator in the server model.
- Recalculation now resets franchise production each round and adds the rounded-up lost production from each externally franchised city to the franchise owner nation.
- The target city's production loss remains unchanged and disclosure-safe projections do not expose hidden city internals.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailFranchise(StealsProductionFromTargetCity|TransfersLostProductionToOwner)' -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail use of the accumulated franchise production inside readiness/material-pool budgeting remains future government/readiness work.
