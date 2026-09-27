# CTP2 Full Retail Tasks 308-311: targeted steal-technology specific chance

## Source notes
- Retail `UnitData::StealTechnology` branches on `whichAdvance < 0` for random steals and uses `SpecificChance` when a concrete advance is requested.
- Retail `Units.txt` contains `StealTechnology { RandomChance ... SpecificChance ... }` for spy units.

## Implementation
- Added optional `advance_id` to the steal-technology command payload and HTTP gateway allow-list.
- Random steals continue to use `special_steal` chance and deterministic sorted eligible random selection.
- Targeted steals validate that the target owns the requested advance and the actor does not, then use `special_steal_specific_chance` and transfer only that requested advance.
- Projection stays safe: legal actions still advertise the generic steal action and do not enumerate hidden target advances.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailStealTechnologySpecificChanceStealsRequestedAdvance|TestTask11StealTechnologySortedRNGFailureCaptureAndDisclosure' -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/http -count=1`
- `go test ./internal/gameplay/visibility -count=1`
- `go test ./internal/gameplay/ai -count=1`
