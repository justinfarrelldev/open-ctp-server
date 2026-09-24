# CTP2 Full Retail Task 76: population-consuming production legality

## Source references
- Retail `BuildingRemovesAPop` units consume one population when produced.
- Such units should not be started by a city without spare population.

## Implementation
- City production validation rejects population-removing unit production when the city is at population one.
- Owner action projection and deterministic AI production selection skip population-consuming units for size-one cities.
- Completion still reduces population for legal production and never drops below one.

## Tests
- `go test ./internal/gameplay/engine -run TestRuntimePopulationRemovingUnitProductionRequiresSparePopulation -count=1`
- `go test ./internal/gameplay/engine -run TestRetailPopulationRemovingUnitsReduceCityPopulation -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/ai ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Full CTP2 city-size rules and specialist assignment around settler-style production still need parity review.
