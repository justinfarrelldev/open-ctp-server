# CTP2 Full Retail Task 183: overseas peace-movement deployment penalty

## Source evidence
- `PlayHap.cpp::CalcPeaceMovement` obtains `AtHomeRadius` and `OverseasCoef` from the player's government.
- It finds each army's nearest owned city and adds `distance - radius` to overseas distance when the army is outside the at-home radius.
- It converts aggregate overseas distance to a per-city happiness penalty.

## Implementation
- Government effects now expose `at_home_radius` and `overseas_coef` from runtime government definitions.
- City happiness now subtracts a deterministic per-city overseas deployment penalty for owned units farther than the at-home radius from the nearest owned city.
- Transported units and unavailable tile/city references are skipped safely.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailOverseasUnitsReduceHappinessOutsideAtHomeRadius -count=1`
- `go test ./internal/gameplay/engine -run TestGovernmentRuntimeYieldPercentEffects -count=1`
- `go test ./internal/gameplay/engine -count=1`
