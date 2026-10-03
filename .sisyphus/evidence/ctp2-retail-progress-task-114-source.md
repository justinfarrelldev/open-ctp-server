# Task 114 source evidence: global/diplomacy wonder effects

## Retail source checks

- Retail parses global/diplomacy wonder fields: `EmbassiesEverywhere`, `CloseEmbassies`, `SpiesEverywhere`, `RevoltingCitiesJoinPlayer`, `EmbassiesEverywhereEvenAtWar`, `IncreaseRegard`, `PreventConversion`, and `ProtectFromBarbarians` (`ctp2_source_code/ctp2_code/gs/newdb/WonderRecord.cpp:206-273`).
- Embassy/contact behavior:
  - `Player.cpp:8479-8492` rejects embassies when the target owns `CloseEmbassies`, grants embassy/contact immediately for the actor's `EmbassiesEverywhereEvenAtWar`, and otherwise applies the non-at-war embassy logic.
  - `Player.cpp:7565-7579`/`8479-8484` and `diplomat.cpp:3879` use close/everywhere wonder helpers in player/diplomacy paths.
- `IncreaseRegard` contributes to AI/diplomatic regard (`diplomat.cpp:1200`) and is also applied on wonder build (`Player.cpp:7667`).
- `SpiesEverywhere` is used to expose city investigation/details without a local spy (`UnitData.cpp:3688`).
- `ProtectFromBarbarians` blocks vandal/barbarian attacks (`armyevent.cpp:675`, `cellunitlist.cpp:531`).
- `RevoltingCitiesJoinPlayer` redirects revolt ownership (`CityData.cpp:767-772`).

## Server coverage

- `CloseEmbassies` removes inbound embassies at wonder completion and suppresses effective/projected embassies.
- `EmbassiesEverywhereEvenAtWar` is imported as the effective-embassy marker and deliberately ignores war state while respecting `CloseEmbassies`.
- `IncreaseRegard` is projected as positive diplomatic regard.
- `SpiesEverywhere`, `ProtectFromBarbarians`, and `RevoltingCitiesJoinPlayer` are executed in investigation, barbarian combat, and incite-revolution flows.
- Added `TestRetailEmbassiesEverywhereEvenAtWarGrantsEffectiveEmbassy` to lock the at-war/close-embassy edge case.

## Validation

- Target GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-114-target-green.txt`
- Visibility GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-114-visibility-green.txt`
- Package GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-114-green.txt`
