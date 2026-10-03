# Task 093 source notes — active-defense enabler gates complete

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/UnitRecord.cpp:676-677,1777-1781` parses `EnableCarrierDefenses` and `ActiveDefenseOnlyWhenCarryingEnablers`; `UnitRecord.h:1261-1262` exposes both flags.
- `ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:1308-1341` implements active-defense eligibility: units with positive `ActiveDefenseRange` must match the attacker's movement domain, and if `ActiveDefenseOnlyWhenCarryingEnablers` is set then the defender only actively defends while carrying cargo whose DB record has `EnableCarrierDefenses`.
- `ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:977-1013` searches nearby active-defense units and runs active-defense triggers during an attack.

Server parity:

- `internal/gameplay/rules/retail_runtime.go` maps retail `ActiveDefenseOnlyWhenCarryingEnablers` to `EffectActiveDefenseRequiresEnabler` and `EnableCarrierDefenses` to `EffectEnableCarrierDefenses`.
- `internal/gameplay/engine/combat.go:548-579` filters active-defense interceptors by range, ownership, domain defend metadata, one-shot state, and `activeDefenseHasRequiredEnabler` cargo checks.
- Active defense is automatic bombard/combat resolution state, not a player-issued legal action; public legal-action planning only exposes the bombard action that may trigger it.

Validation:

- `go test ./internal/gameplay/engine -run 'TestRetailActiveDefenseRequiresCarrierEnabler|TestRetailActiveDefenseRangeIsMeasuredToBombarder|TestRetailActiveDefenseFiresOncePerTurn|TestRetailActiveDefenseUseResetsOnTurnRefresh|TestRetailActiveDefenseInterceptorsSortByAttackStrength' -count=1`
