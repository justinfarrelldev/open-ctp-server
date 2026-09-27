# CTP2 Full Retail Tasks 258-260: NoSlaves and slaver production exclusion

## Retail source
- `ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp` and `CityData.cpp` reject slaver-producing records when an owned `NoSlaves` unit exists, and reject `NoSlaves` units when owned slave-raid units/slaves exist.

## Implementation
- Engine production rejects slaver units when the actor owns a `no_slaves` unit.
- Engine production rejects `no_slaves` units when the actor owns a slave-raid unit.
- Safe projections suppress these mutually-exclusive production actions and continue exposing the next legal fallback production action.

## Tests
- `go test ./internal/gameplay/engine -run TestRuntimeNoSlavesAndSlaverProductionExcludeEachOther -count=1`
- `go test ./internal/gameplay/visibility -run TestProjectionSuppressesNoSlavesAndSlaverExclusiveProduction -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
