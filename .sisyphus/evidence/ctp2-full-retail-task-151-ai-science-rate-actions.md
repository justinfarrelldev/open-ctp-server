# CTP2 Full Retail Task 151: AI uses projected science-rate actions

## Implementation
- Science-rate legal-action projection now treats an unset nation science percent as the default 100%, avoiding redundant `set-science-rate` 100 actions.
- AI planning consumes projected `set-science-rate` legal actions, but only chooses the 100% science action so it does not voluntarily lower science without a fuller tax/luxury strategy.
- The AI copies the projected payload exactly and does not synthesize hidden government/economy commands.

## Tests
- `go test ./internal/gameplay/visibility -run TestScienceRateLegalActionsTreatUnsetAsDefaultHundred -count=1`
- `go test ./internal/gameplay/ai -run TestPlannerUsesProjectedScienceRateAction -count=1`
- `go test ./internal/gameplay/ai -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Full AI slider optimization across tax/luxury/science remains future parity work.
