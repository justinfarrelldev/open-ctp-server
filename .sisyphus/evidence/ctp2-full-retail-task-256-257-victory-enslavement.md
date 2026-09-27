# CTP2 Full Retail Tasks 256-257: victory enslavement combat resolution

## Retail source
- `ctp2_source_code/ctp2_code/gs/gameobj/armyevent.cpp` calls `DoVictoryEnslavement` for the surviving army after combat.
- `ctp2_source_code/ctp2_code/gs/world/cellunitlist.cpp` applies `VictoryEnslavement` once from a surviving marked unit and chooses an owner slave city.
- `ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp` picks the nearest eligible slave city, falling back to nearest city.

## Implementation
- Successful combat now applies `victory_enslavement` for either attacking or defending surviving armies.
- The first surviving marked unit increases the nearest owned city population by one, matching the modeled population/slavery abstraction used by existing slave-raid systems.
- Combat event payload remains safe and aggregate-only by adding `enslaved_population` count.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailVictoryEnslavementAddsPopulationToNearestWinnerCity -count=1`
- `go test ./internal/gameplay/engine -count=1`
