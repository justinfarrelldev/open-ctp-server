# CTP2 Full Retail Task 145: combat veteran promotion

## Implementation
- Decisive combat now promotes surviving units in the winning army to `veteran`.
- Veteran state persists on canonical units and feeds the elite special-action chance path implemented in Task 144.
- Retreated/non-decisive combats do not currently promote either side.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailCombatPromotesSurvivingWinnersToVeteran -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Precise CTP2 multi-level experience and promotion probabilities remain future combat parity work.
