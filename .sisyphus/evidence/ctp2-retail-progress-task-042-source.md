# Task 042: retail Advertise unit action

## Retail source/data checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:3480-3520`
  - `ArmyData::CanAdvertise` requires a unit record with `Advertise` and available special action.
  - `ArmyData::Advertise` requires an adjacent foreign city, rejects own cities, emits attacker/victim messages, and delegates to `CauseUnhappiness(point, uindex)`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:1710-1816` (CauseUnhappiness flow previously implemented)
  - `CauseUnhappiness` uses the nested `CauseUnhappiness` data (`Chance`, `Timer`, `Amount`) to apply a timed unhappiness effect.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Units.txt:940-1001`
  - `UNIT_CORPORATE_BRANCH` has `Advertise` and nested `CauseUnhappiness { Chance 1 Timer 1 Amount 5 ... }` data.

## Implementation notes

- Added `CommandAdvertise` / `EventAdvertised`.
- `advertise` requires both the retail `EffectAdvertise` marker and the already-imported `EffectSpecialCauseUnhappiness` data, then reuses the retail unhappiness attack path so amount/timer/chance semantics remain shared.
- Public legal actions now project `advertise` for retail advertiser units instead of the generic `cause-unhappiness` route.
- HTTP gateway accepts `/game/{id}/advertise` with only `unit_id` and `city_id` forwarded to the canonical command.

## Red/green evidence

- Red engine test: `.sisyphus/evidence/ctp2-retail-progress-task-042-red.txt`
- Red visibility test: `.sisyphus/evidence/ctp2-retail-progress-task-042-red-visibility.txt`
- Red HTTP test: `.sisyphus/evidence/ctp2-retail-progress-task-042-red-http.txt`
- Green targeted tests:
  - `go test ./internal/gameplay/engine -run TestAdvertise -count=1`
  - `go test ./internal/gameplay/visibility -run TestAdvertiseLegalActionUsesRetailAdvertiseRoute -count=1`
  - `go test ./internal/gameplay/http -run TestHandlerDelegatesAdvertiseRoute -count=1`
