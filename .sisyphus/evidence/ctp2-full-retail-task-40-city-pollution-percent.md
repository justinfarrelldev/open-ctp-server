# CTP2 Full Retail Task 40: city pollution percentage effects

## Source references
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:1041-1125` (`CityData::CalcPollution`) computes population and industrial pollution, applies owner government pollution coefficient, then adds effective building `PopulationPollutionPercent`, `ProductionPollutionPercent`, and flat pollution contributions.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/building.cdb:55-56` defines `PopulationPollutionPercent` and `ProductionPollutionPercent` building fields.
- Released retail data uses these fields in `buildings.txt` for pollution reducers including `IMPROVE_ECO_TRANSIT` and `IMPROVE_MATTER_DECOMPILER`.

## Implementation
- Runtime projection maps retail building `PopulationPollutionPercent` and `ProductionPollutionPercent` to safe public effects `pollution_population_percent` and `pollution_production_percent`.
- Round recalculation derives bounded city pollution from positive population plus current production yield, sums percentage effects per CTP2's additive percentage model, clamps negative buckets to zero, and writes the result to the city tile.
- City pollution integration is enabled only when runtime pollution triggers are present, preserving legacy synthetic scenarios while retail rules remain fully opted into pollution turn integration.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailCityPollutionUsesPopulationProductionAndReducers|TestCityPollutionRequiresRetailPollutionTriggers' -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `GOMAXPROCS=4 go test ./internal/e2e/fullgame -run TestTaskFiveExecutionSemanticsRemainGreen -count=1`
- `go test ./internal/e2e/fullgame -count=1`
- `go test ./internal/e2e/freeze -count=1`

## Remaining gaps
- This is still a bounded model: difficulty start levels/ratios, government `PollutionCoef`, wonder pollution effects, and local pollution happiness are not fully modeled yet.
