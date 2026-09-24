# CTP2 Full Retail Task 143: AI uses projected unit-target special actions

## Implementation
- AI planning now consumes projection-authorized unit-target `expel` and `sue` legal actions.
- Candidates copy the projection payload directly after validating `unit_id` and `target_unit_id`, so AI cannot synthesize hidden targets.
- Duplicate command/unit/target tuples are collapsed deterministically before candidate sorting.

## Tests
- `go test ./internal/gameplay/ai -run 'TestPlannerUsesProjected(UnitSpecialTargetActions|EstablishContactAction)' -count=1`
- `go test ./internal/gameplay/ai -count=1`

## Remaining gaps
- Broader AI use of city-target special actions remains future parity work.
