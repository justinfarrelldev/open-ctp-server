# CTP2 Full Retail Task 110: protect-from-barbarians combat effect

## Source references
- `armyevent.cpp` cancels combat when the attacker owner is `PLAYER_INDEX_VANDALS` and the defender owner has a `ProtectFromBarbarians` wonder.
- Retail `Wonder.txt` includes `ProtectFromBarbarians` on Labyrinth.

## Implementation
- Combat legality now rejects barbarian/vandal attacks when the defender owner has an active `protect_from_barbarians` wonder.
- Barbarian detection recognizes canonical barbarian player IDs and `CIVILIZATION_BARBARIAN`/barbarian nation IDs.
- The rejection occurs before combat RNG resolution.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailProtectFromBarbariansBlocksBarbarianCombat -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Startup currently does not persist the selected civilization record into gameplay player state; this task uses available player/nation IDs until country metadata is carried into canonical gameplay state.
