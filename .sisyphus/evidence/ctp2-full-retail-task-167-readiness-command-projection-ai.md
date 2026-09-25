# CTP2 Full Retail Task 167: readiness command, projection, and AI

## Implementation
- Added canonical `Nation.Readiness` plus `set-readiness` command and `government.readiness_changed` event.
- Readiness payloads are validated to `peace`, `alert`, or `war`; unsupported levels fail closed.
- Owner projections include readiness safely and legal actions offer non-current readiness levels.
- HTTP command routing supports `set-readiness`.
- AI consumes projected `set-readiness` actions and currently only selects `war`, matching the existing conservative military strategy.

## Tests
- `go test ./internal/gameplay/engine -run TestSetReadiness -count=1`
- `go test ./internal/gameplay/visibility -run TestReadiness -count=1`
- `go test ./internal/gameplay/ai -run TestPlannerUsesProjectedReadinessWarAction -count=1`
- `go test ./internal/gameplay/http -count=1`
