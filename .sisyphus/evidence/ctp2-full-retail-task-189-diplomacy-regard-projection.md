# CTP2 Full Retail Task 189: diplomacy regard projection

## Source evidence
- `Regard.cpp::GetUpdatedRegard` maps human-visible diplomatic state to regard: war becomes `HOTWAR`, neutral/ceasefire remain neutral, allied becomes love.
- The server already tracks war edges, contacted relations, and `union` as the closest current allied treaty representation.

## Implementation
- Safe diplomacy projection now includes derived `regard` for contacted visible relations.
- War policy edges project `hotwar`; union treaties project `love`; all other contacted relations currently project `neutral`.
- This is derived only from already-visible relation/policy data and does not expose hidden relations.

## Tests
- `go test ./internal/gameplay/visibility -run TestDiplomacyProjectionIncludesRetailRegard -count=1`
- `go test ./internal/gameplay/visibility -count=1`
