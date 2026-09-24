# CTP2 Full Retail Task 121: slave-raid chances

## Implementation
- Slave raid and settler-slave-raid actions now honor positive effect values as deterministic percent success chances.
- Zero-value legacy/synthetic raid effects remain 100% success for existing fixtures.
- Chance failures spend movement and emit `city.slave_raid_failed` without reducing city population.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailSlaveRaidsReducePopulationAndHonorProtection -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail slave objects, death/capture chances, timers, and diplomacy regard effects remain future work.
