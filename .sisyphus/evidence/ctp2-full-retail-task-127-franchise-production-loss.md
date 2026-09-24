# CTP2 Full Retail Task 127: franchise production loss

## Source notes
- CTP2 `CityData::ComputeGrossProduction` computes `franchise_loss = ceil(net_production * g_theConstDB->GetFranchiseEffect())` when a city is franchised.
- Retail `Const.txt` has `FRANCHISE_EFFECT 0.1`.

## Implementation
- Cities with an external `franchise:<owner>` marker now lose 10% of post-crime production, rounded up.
- The loss is applied to the target city's yield and clamped by the existing non-negative yield floor.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailFranchiseStealsProductionFromTargetCity -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Transfer of stolen franchise production to the franchise owner/readiness budget and exact franchise duration/sue flow remain future work.
