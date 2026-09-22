# Task 5 pinned source review

- Reviewed UTC: 2026-08-11
- Pinned source: `civctp2/civctp2@0668dd199c2519ff8ab7d2871e7c7a89e21c8e52`
- Relevant paths/symbols:
  - `ctp2_code/gs/gameobj/PlayerTurn.cpp`: `Player::BeginTurn`
  - `ctp2_code/gs/gameobj/Player.cpp`: `Player::ResetAllMovement`
  - `ctp2_code/gs/gameobj/UnitData.cpp`: `UnitData::ResetMovement`, `UnitData::ResetCityOwner`
  - `ctp2_code/gs/events/armyevent.cpp`: attacker victory movement and `CaptureCityEvent`
  - `ctp2_code/gs/events/CityEvent.cpp`: city capture event processing
  - `ctp2_code/gs/gameobj/CityData.cpp`: `CityData::ResetCityOwner`
  - `ctp2_code/gs/gameobj/Vision.cpp`: visibility consequences of ownership and movement
- Observed behavior: a victorious attacker enters the defended city tile and city ownership/visibility transfer is performed as part of the capture sequence before ordinary public consequences continue. The pinned source has no separate post-victory "legal frontier" requirement.
- Server correspondence: `internal/gameplay/engine/applyCombat` moves a victorious army to the target tile, calls `captureCityAt`, and returns the resulting city ID in `CombatResult.CapturedCityID`; `internal/gameplay/service.Service.executeControllerCommand` refreshes visibility and exposes that result through the actor's public response; `internal/e2e/fullgame.HTTPExplorationTransport.Combat` forwards it in `CombatTransition.Result`.
- Implementation consequence: treat a non-empty public `CombatTransition.Result.CapturedCityID` only as acknowledgement that authoritative capture already succeeded. Do not query the ID or use it for target selection. Preserve the existing projection-only `captureVisibleForeignCity` fallback when the public result has an empty captured ID.
