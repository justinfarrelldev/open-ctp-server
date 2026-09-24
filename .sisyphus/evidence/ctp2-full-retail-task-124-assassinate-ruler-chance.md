# CTP2 Full Retail Task 124: assassinate-ruler chance

## Implementation
- `special_assassinate_ruler` positive effect values now act as deterministic percent success chances.
- Failed chance rolls spend movement, do not destabilize the target nation, and emit disclosure-safe `nation.ruler_assassination_failed`.
- Zero-value legacy/synthetic effects remain 100% success for existing fixtures.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailAssassinateRulerDestabilizesTargetNation -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail elite chance, death chance, and diplomatic regard changes remain future work.
