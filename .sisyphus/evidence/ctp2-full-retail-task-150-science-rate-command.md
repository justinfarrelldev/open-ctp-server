# CTP2 Full Retail Task 150: science-rate command and projection

## Implementation
- Added `set-science-rate` command with validation for explicit `science_percent` values from 0 through 100.
- The command persists an explicit pointer value so zero science is distinguishable from an unset/default rate.
- Added `government.science_rate_changed` events.
- Owner projections now advertise safe preset `set-science-rate` legal actions (0, 50, 100) without exposing hidden state.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailSetScienceRateCommandPersistsExplicitZero -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Tax/luxury sliders, wages/rations/workday, and AI slider optimization remain future parity work.
