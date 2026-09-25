# CTP2 Full Retail Task 171: war discontent happiness

## Source evidence
- `PlayHap.cpp` counts attacking military units outside cities/forts and computes war discontent when they exceed `WarDiscontentMaxUnits`.
- `Happy.cpp::CalcPeaceMovement` subtracts global war discontent from each city happiness calculation.

## Implementation
- Government `WarDiscontentMaxUnits` and `WarDiscontentPerUnit` now become active nation effects.
- City happiness now subtracts per-city war discontent for owned attack-capable units outside owned city tiles.
- The penalty is deterministic and divided across owned cities with integer-safe rounding.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailWarDiscontentReducesCityHappiness -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Fort exemptions and overseas/home peace-movement defeat terms remain future parity work.
