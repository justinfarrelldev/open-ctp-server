# CTP2 Full Retail Task 148: AI uses projected city-target special actions

## Implementation
- AI planning now consumes projection-authorized city-target special actions, including franchise, conversion, investigation, terrorism, slave raids, underground railway, injunction/reform, assassinate ruler, and plant nuke.
- Candidates validate and copy projected `unit_id`/`city_id` payloads directly, preserving the projection trust boundary and avoiding hidden target synthesis.
- Duplicate command/unit/city tuples are collapsed deterministically before candidate sorting.

## Tests
- `go test ./internal/gameplay/ai -run 'TestPlannerUsesProjected(CitySpecialActions|UnitSpecialTargetActions)' -count=1`
- `go test ./internal/gameplay/ai -count=1`

## Remaining gaps
- Strategic valuation and restraint for late-game special actions remains future AI parity work.
