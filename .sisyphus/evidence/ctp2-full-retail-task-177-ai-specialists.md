# CTP2 Full Retail Task 177: AI specialist assignment

## Implementation
- AI now consumes projected `set-specialists` legal actions.
- For projected owner cities with low happiness, the AI creates a legal specialist payload that adds one entertainer when population capacity is available.
- The AI uses only projected city data and projected legal-action metadata; it does not inspect hidden canonical state.

## Tests
- `go test ./internal/gameplay/ai -run TestPlannerUsesProjectedSpecialistActionForUnhappyCity -count=1`
- `go test ./internal/gameplay/ai -count=1`
