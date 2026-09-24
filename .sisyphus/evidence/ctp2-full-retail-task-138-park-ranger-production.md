# CTP2 Full Retail Task 138: park-ranger production gating

## Source evidence
- `CityData.cpp` and `Player.cpp` reject units with `CreateParks` unless `wonderutil_GetParkRangersEnabled(...)` is true.
- `wonderutil_GetParkRangersEnabled` is driven by retail wonder flag `EnableParkRangers`.

## Implementation
- Unit production now detects `special_create_park` capabilities from runtime or explicit item effects.
- Park-ranger / create-park-capable unit production is rejected unless the actor has an active `enable_park_rangers` wonder effect.
- The create-park action itself remains executable for already-existing units, matching retail's production-gate behavior.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailParkRangerProductionRequiresEnablingWonder -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Create-park diplomacy regard costs remain future diplomacy work.
