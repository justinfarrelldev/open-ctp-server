# CTP2 Full Retail Task 156: empire-distance happiness penalty

## Implementation
- Retail government empire-distance settings now become nation effects when changing government.
- City happiness applies a bounded distance-to-capital penalty using Manhattan tile distance scaled to retail movement-cost units, honoring min/max empire distance settings.
- `DecreaseEmpireSize` wonder effects reduce the penalty percentage.
- Capital detection uses active city effects, so palaces/capitol effects and obsolete-safe city effect calculation remain authoritative.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailEmpireDistanceGovernmentPenaltyReducesHappiness -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Path-cost exactness, difficulty distance adjustment, and conquest/war-distress happiness remain future parity work.
