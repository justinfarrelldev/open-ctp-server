# CTP2 full-retail parity task 15 unit ranged/bombard projection evidence

Scope: project retail ranged/bombard unit stats into runtime and produced units.

Original source/data checked:
- `ctp2_code/gs/newdb/unit.cdb` defines `ZBRangeAttack`, `BombRounds`, and `BombardRange` on units.
- `ctp2_code/gs/gameobj/bldque.cpp::BuildFrontUnit` creates units from the DB record at production completion; engine combat already has a ranged combat phase.
- Retail data example verified: `UNIT_AIRCRAFT_CARRIER` has `ZBRangeAttack 25`, `BombRounds 10`, and `BombardRange 1`.

Implemented server-compatible slice:
- Runtime unit definitions carry ranged attack plus bombard range/round metadata.
- Retail projection maps `ZBRangeAttack`, `BombardRange`, and `BombRounds`.
- Synthetic compile accepts optional `Ranged`/`ZBRangeAttack`, `BombardRange`, and `BombRounds` tags.
- Produced units now copy the runtime ranged attack value into authoritative unit state, enabling the existing ranged combat phase.

Validation commands:
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run ProductionCompletionCopiesRuntimeRangedStats -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/game -count=1`
