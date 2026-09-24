# CTP2 Full Retail Task 106: temporary full happiness wonder effect

## Source references
- `Player.cpp` applies `TemporaryFullHappiness` on wonder completion by setting full-happiness turns on all owner cities.
- Retail `Wonder.txt` uses `TemporaryFullHappiness 5` on Ramayana.

## Implementation
- Wonder completion now adds expiring city-local `temporary_full_happiness` markers to all owner cities.
- City happiness is floored to full happiness (`100`) while the marker remains active.
- Markers decrement deterministically during recalculation and are removed when expired.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailTemporaryFullHappinessWonderAddsExpiringCityMarkers -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail citizen-level happiness classes and riot interactions remain broader future work.
