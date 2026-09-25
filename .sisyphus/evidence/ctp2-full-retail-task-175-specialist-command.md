# CTP2 Full Retail Task 175: specialist assignment command/API

## Implementation
- Added `set-specialists` command and `city.specialists_changed` event.
- Specialist payloads validate city ownership, known retail specialist IDs, nonnegative counts, and total specialists within city population.
- HTTP command routing accepts `set-specialists` with `city_id` and `specialists` fields.
- Owner projections expose specialist assignments and legal actions for specialist updates.

## Tests
- `go test ./internal/gameplay/engine -run TestSetSpecialists -count=1`
- `go test ./internal/gameplay/visibility -run TestSpecialists -count=1`
- `go test ./internal/gameplay/http -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
