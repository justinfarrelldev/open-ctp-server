# CTP2 Full Retail Task 90: active-defense bombard interception

## Source references
- Retail `ActiveDefenseRange` gives units an interception/defensive-fire role against bombardment.
- Active defense is resolved as a defensive combat layer distinct from the bombarding unit's normal rounds.

## Implementation
- Bombard resolution now searches for deterministic same-owner active-defense interceptors around the target tile.
- The first eligible interceptor fires before bombard rounds using deterministic combat RNG.
- If active defense destroys the bombarder, bombard rounds do not continue and the bombarder is removed.
- Bombard damage calculations now guard against zero armor.
- Bombard events include an `active_defense_hits` count when interception hits.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailActiveDefenseInterceptsBombardment -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Full retail multi-interceptor selection, counter-bombard, air/naval/domain masks, ammunition/round accounting, and diplomacy messaging remain to be matched.
