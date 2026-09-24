# CTP2 Full Retail Task 86: assassinate ruler special action

## Source references
- Retail unit field `AssasinateRuler` marks units able to destabilize enemy governments.
- Ruler assassination is a nation-level special action rather than ordinary combat.

## Implementation
- Added `assassinate-ruler` command and disclosure-safe event.
- Capable adjacent owned units spend movement against enemy cities.
- Successful assassination clears the target nation's active government, resets research progress, and records a nation effect marker.
- State validation now accepts the ruler-assassinated effect marker.
- Owner legal-action projections expose visible adjacent assassination targets for capable units.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailAssassinateRulerDestabilizesTargetNation -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail anarchy duration, assassination chance/counter-intelligence, diplomatic effects, and government recovery logic remain to be implemented.
