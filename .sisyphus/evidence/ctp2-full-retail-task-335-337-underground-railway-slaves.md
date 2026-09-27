# CTP2 Full Retail Tasks 335-337: underground railway slave transfer

## Source notes
- Retail `ArmyData::UndergroundRailway` requires the target city to have slaves.
- Retail `UndergroundRailwayUnitEvent` removes one slave from the target city and adds one population to the nearest city owned by the abolitionist owner.

## Implementation
- Underground railway now requires `POP_SLAVE` in the target city instead of generic population.
- Success consumes one target slave specialist without reducing target city population.
- Success adds one population to the nearest owned city for the acting player and keeps the aggregate freed-population effect/event.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailUndergroundRailwayFreesSlaveToNearestHomeCity|TestTask11RetailSpecialDeathChanceRemovesUnit' -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
- `go test ./internal/gameplay/ai -count=1`
