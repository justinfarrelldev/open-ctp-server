# CTP2 Full Retail Task 93: underground railway special action

## Source references
- Retail unit field `UndergroundRailway` marks units able to free population/slaves from enemy cities.
- Retail wonder field `FreeSlaves` depends on slave/freed-population semantics.

## Implementation
- Added `underground-railway` command and disclosure-safe event.
- Capable adjacent owned units spend movement to free population from enemy cities above size one.
- The target city loses one population and the actor nation records a `freed_population` counter effect.
- State validation accepts the freed-population nation effect.
- Owner legal-action projections expose visible adjacent underground-railway targets for capable units.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailUndergroundRailwayFreesPopulation -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Full retail slave objects, abolition/wonder interactions, diplomacy, and exact success/capture rules remain to be implemented.
