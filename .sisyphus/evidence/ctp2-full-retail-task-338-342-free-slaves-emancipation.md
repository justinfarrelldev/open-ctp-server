# CTP2 Full Retail Tasks 338-342: free-slaves emancipation

## Source notes
- Retail `Player::Emancipate` calls `CityData::FreeSlaves`, removes illegal build items, applies emancipation unhappiness timers to cities where slaves were freed, and kills slave-raid / settler-slave-raid / slave-uprising units.
- Retail `Const.txt` defines `EMANCIPATION_UNHAPPINESS_TURNS` and `EMANCIPATION_UNHAPPINESS_AMOUNT`.

## Implementation
- Imported emancipation unhappiness duration/amount into runtime constants.
- `FreeSlaves` wonder completion now removes `POP_SLAVE` specialists from all cities.
- Non-owner cities that lose slaves receive the retail emancipation unhappiness timer.
- Emancipation now removes slave-uprising units in addition to slave-raid units.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestRetailFreeSlavesWonderEmancipatesCitiesAndSlaverUnitsEverywhere -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/visibility -count=1`
