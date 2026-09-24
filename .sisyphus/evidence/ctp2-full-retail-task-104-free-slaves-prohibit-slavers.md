# CTP2 Full Retail Task 104: free-slaves and prohibit-slavers wonder effects

## Source references
- `Player.cpp` applies `FreeSlaves` on wonder completion by freeing slaves in cities and removing slave-raiding / slave-uprising units.
- `Player::Emancipate` removes slave-capable units and removes illegal build-queue items.
- Retail `Wonder.txt` gives Emancipation both `FreeSlaves` and `ProhibitSlavers`.

## Implementation
- Completing a `free_slaves` wonder removes current slave-raiding and settler-slave-raiding units from all players using the canonical unit-removal path.
- `prohibit_slavers` blocks future slave-raid and settler-slave-raid special commands for the affected player.
- `prohibit_slavers` also blocks production queuing of units whose runtime/unit effects include slave-raiding capabilities.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetail(FreeSlaves|ProhibitSlavers)' -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- No dedicated slave-population counter exists yet, so city slave-labor accounting and emancipation unhappiness timers remain future work.
