# CTP2 Full Retail Task 98: crime yield loss

## Source references
- Retail crime reduces city economic output; crime reducers are separate city/building/government effects.
- Crime should not directly consume food but should reduce productive/economic yields.

## Implementation
- City yield calculation now applies positive city `crime_percent` as a production and gold loss.
- Food is left unchanged by crime loss.
- Crime loss is clamped to the 0..100 range and applied after building upkeep.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailCrimePercentReducesProductionAndGold -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Full retail crime generation from empire size/government/difficulty, unrest interactions, and per-city crime reducers still need deeper parity.
