# CTP2 Full Retail Task 141: visible foreign contact projection

## Implementation
- Viewer legal-action projection now offers `establish-contact` when a visible city/unit belongs to a foreign player without existing contact.
- The action payload exposes only `target_player_id`; it is derived from already-visible projected entities and does not leak hidden players.
- Existing contact suppresses the action and normal agreement proposal/response projections continue to use contacted relations.

## Tests
- `go test ./internal/gameplay/visibility -run TestRetailVisibleForeignOwnerProjectsEstablishContact -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Automatic contact establishment during movement/reveal remains future diplomacy parity work.
