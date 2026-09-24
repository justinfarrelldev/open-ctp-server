# CTP2 Full Retail Task 152: government social slider fields

## Implementation
- Runtime government definitions now preserve retail workday, wages, and rations expectation/coefficient fields.
- Retail runtime import maps these directly from `govern.txt` `IntValues`/`FloatValues`.
- Synthetic/manual runtime compilation can parse the same optional float/int fields.

## Tests
- `go test ./internal/gameplay/rules -count=1`

## Remaining gaps
- Applying workday/wages/rations to happiness/economy and exposing full slider commands remains future parity work.
