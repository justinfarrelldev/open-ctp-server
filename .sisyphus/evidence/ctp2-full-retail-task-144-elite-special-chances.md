# CTP2 Full Retail Task 144: elite special-action chances

## Implementation
- Retail unit `EliteChance` subfields now import as `*_elite_chance` runtime effects for grouped special actions.
- Canonical units now persist a `veteran` flag.
- Veteran special-action attempts prefer the corresponding elite chance when present; normal units continue to use the normal chance.
- Investigate-city and assassinate-ruler now share the city special-action chance path, including existing watchful modifiers.
- Owner projections include the safe `veteran` flag for the viewer's own units only.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailVeteranSpecialActionUsesEliteChance -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeIncludesCompleteReleasedCatalog -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Promotion/veterancy earning from combat remains future combat parity work.
