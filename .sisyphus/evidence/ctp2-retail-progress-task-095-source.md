# Task 095 source notes — IsFlanker tactical targeting complete

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/UnitRecord.cpp:572,1437-1438` parses `IsFlanker`; `UnitRecord.h:969` exposes `GetIsFlanker()`.
- `ctp2_source_code/ctp2_code/gs/gameobj/CTP2Combat.cpp:145-156` classifies a non-air/non-naval flanking unit as `UNIT_TYPE_FLANKER`.
- `ctp2_source_code/ctp2_code/gs/gameobj/CTP2Combat.cpp:303-343` implements flanker targeting: if the same row target is inactive, flankers scan adjacent rows outward in the same column and attack the nearest active row; non-flankers do not retarget.
- `ctp2_source_code/ctp2_code/gs/outcom/C3UnitDB.cpp:710-713` exposes the flanker flag to AI/outcom consumers, and `slicfunc.cpp:806-814` exposes it to SLIC.

Server parity:

- `internal/gameplay/rules/retail_runtime.go` maps retail `IsFlanker` into `EffectIsFlanker`.
- `internal/gameplay/engine/combat.go` uses `EffectIsFlanker` during grouped combat targeting so non-flankers do not retarget inactive rows and flankers strike the nearest occupied row.

Validation:

- `go test ./internal/gameplay/engine -run 'TestRetailNonFlankerMeleeDoesNotRetargetEmptyRows|TestRetailFlankerMeleeTargetsNearestOccupiedRow' -count=1`
