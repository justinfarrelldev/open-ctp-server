# CTP2 Full Retail Task 155: too-many-cities happiness penalty

## Implementation
- Retail government `TooManyCitiesThreshold` and `TooManyCitiesCoefficient` now become nation effects when changing government.
- City happiness applies the retail-style penalty when the owner has more populated cities than the threshold: `ceil((city_count - threshold) * coefficient)`.
- The effect applies through active nation state, so recalculation and projections use canonical city happiness only after state mutation.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailTooManyCitiesGovernmentPenaltyReducesHappiness -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Empire-distance and conquest/war-distress happiness are still future government parity work.
