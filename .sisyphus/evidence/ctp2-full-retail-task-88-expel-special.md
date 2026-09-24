# CTP2 Full Retail Task 88: expel special action

## Source references
- Retail unit field `CanExpel` marks units able to expel other units.
- Expulsion is a special displacement action separate from combat.

## Implementation
- Added `expel` command and disclosure-safe event.
- Capable owned units can expel adjacent visible enemy units, spending movement.
- The expelled unit moves deterministically to an adjacent open, unoccupied, non-city tile, preferring its owner's territory.
- Expelled units lose fortification.
- Owner legal-action projections expose adjacent visible expel targets for capable units.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailExpelMovesAdjacentEnemyUnit -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail expel legality, diplomacy/court outcomes, transport/stack/domain handling, and failure chances remain to be implemented.
