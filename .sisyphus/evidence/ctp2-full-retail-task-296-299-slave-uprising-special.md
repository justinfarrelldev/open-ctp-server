# CTP2 Full Retail Tasks 296-299: slave-uprising special execution

## Implementation
- Added `slave-uprising` as an executable special-action command for retail `SlaveUprising` units.
- Execution requires target-city slaves, spends special movement, consumes one `POP_SLAVE`, and marks the city rioting as the current safe uprising abstraction.
- `POP_SLAVE` is now accepted as a canonical city specialist assignment for imported retail slave mechanics.
- Safe projections, HTTP routes, and AI scoring expose/use the action only from owned capable units.

## Source notes
- Retail `ArmyData::InciteUprising` requires slaves in the target city and schedules `GEV_InciteUprisingUnit`, which invokes `CityData::DoUprising`. Full rebel-army/civilisation creation remains future parity work; the current implementation applies a deterministic safe city-riot abstraction.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailSlaveUprisingConsumesSlaveAndRiotMarksCity -count=1`
- `go test ./internal/gameplay/visibility -run TestTask12ProjectionIncludesRulesAndLegalActionMetadata -count=1`
- `go test ./internal/gameplay/http -count=1`
- `go test ./internal/gameplay/ai -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
