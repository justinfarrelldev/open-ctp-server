# Task 040: retail ThrowParty unit action

## Retail source/data checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:2633-2676`
  - `ArmyData::ThrowParty` charges `g_theConstDB->GetMaximumPartyCost()` (local fallback `100`), requires a `ThrowParty` unit, adjacent foreign city, no declared-war agreement, and `Diplomat::ReadyToParty()`, then enqueues `GEV_ThrowPartyUnit` and subtracts gold.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:4155-4167`
  - `UnitData::ThrowParty` requires the unit record `ThrowParty` flag and enqueues `GEV_ThrowParty` for actor and target owner.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/PlayerEvent.cpp:614-621`
  - `ThrowPartyEvent` dispatches to the target player's `Diplomat::ThrowParty(owner)`.
- `ctp2/ctp2_source_code/ctp2_code/ai/diplomacy/diplomat.cpp:5951-5978`
  - `Diplomat::ThrowParty` logs the hold-reception regard bonus and sets `m_lastParty` to the current round; `ReadyToParty` allows another party only when `m_lastParty + 10 < current round`.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Units.txt:1288,1515`
  - Retail units use the `ThrowParty` flag imported as canonical/runtime `throw_party` marker effects.

## Implementation notes

- Added `CommandThrowParty` / `EventPartyThrown`.
- `throwParty` now executes the retail action against adjacent visible/legal foreign cities:
  - requires a unit with canonical `EffectThrowParty`, one movement/special action, and 100 gold;
  - rejects own-city targets through shared `specialTarget` and rejects active war;
  - enforces target diplomacy reception cooldown with canonical `Nation.LastPartyTurn`;
  - deducts gold, spends movement, sets contact, and adds a 10-point timed directional diplomacy regard bonus from the target player toward the actor (matching `Diplomat::GetDiplomat(target).ThrowParty(actor)`).
- Round advancement decrements timed diplomacy regard and clears the bonus when expired.
- Public projection maps positive regard bonus to `friendly` and exposes `throw-party` legal actions only from public gold + public diplomacy relation state.
- HTTP gateway accepts `/game/{id}/throw-party` with `unit_id` and `city_id` only.

## Red/green evidence

- Red engine test: `.sisyphus/evidence/ctp2-retail-progress-task-040-red.txt`
- Red visibility test: `.sisyphus/evidence/ctp2-retail-progress-task-040-red-visibility.txt`
- Red HTTP test: `.sisyphus/evidence/ctp2-retail-progress-task-040-red-http.txt`
- Red directional-regard test: `.sisyphus/evidence/ctp2-retail-progress-task-040-red-directional-regard.txt`
- Green targeted tests run after implementation:
  - `go test ./internal/gameplay/engine -run TestThrowParty -count=1`
  - `go test ./internal/gameplay/visibility -run TestThrowParty -count=1`
  - `go test ./internal/gameplay/http -run TestHandlerDelegatesThrowPartyRoute -count=1`
- Package validation:
  - `go test ./internal/gameplay/... -count=1`
  - `go test ./internal/e2e/freeze -count=1`
