# CTP2 Full Retail Task 159: government crime fields

## Implementation
- Runtime governments now preserve retail `CrimeCoef`, `CrimeOffset`, and `PollutionUnhappyCoef` values from `govern.txt` and manual definitions.
- Government changes add active nation effects for crime offset/coefficient.
- City crime now computes the retail happiness-derived government crime component: `ceil((crime_offset - happiness) * crime_coef)`, capped only at 100 so negative building/wonder crime reducers remain possible.

## Tests
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -run TestRetailGovernmentCrimeCoefComputesCrimeFromHappiness -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Pollution-unhappiness application remains future parity work.
