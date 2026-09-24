# CTP2 Full Retail Task 153: government readiness/support fields

## Implementation
- Runtime government definitions now preserve retail readiness/support fields: support/defense coefficients, professional unit count, readiness transition turns, and peace/alert/war readiness coefficients and hit-point multipliers.
- Retail import maps these values directly from `govern.txt`.
- Synthetic/manual runtime compilation accepts the same optional fields.

## Tests
- `go test ./internal/gameplay/rules -count=1`

## Remaining gaps
- Applying readiness/support costs and readiness combat HP modifiers remains future parity work.
