# Task 094 source notes — advance Age setup, end-age research, and age transitions

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/AdvanceRecord.cpp:466-469` parses each advance `Age` as an Age DB item, and `AdvanceRecord.cpp:621-624` exposes the referenced `AgeRecord`.
- `ctp2_source_code/ctp2_code/gs/newdb/AgeRecord.cpp:133-148` parses `Age`, `StartingRound`, `NeedAdvances`, and `SlicObject` from `age.txt`.
- `ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:549-560` grants all advances whose `GetAgeIndex()` is before the configured starting age during player initialization.
- `ctp2_source_code/ctp2_code/gs/gameobj/Advances.cpp:493-496` suppresses research for advances whose age is greater than the configured ending age.
- `ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:10046-10058` enters a new age when a newly obtained advance belongs to a later age and the player has at least that age record's `NeedAdvances` count.
- `ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:10553-10572` stores the player's new age and runs retail presentation/event work (`SlicObject`, city sprites, event tracker).
- `ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:562-572` also uses starting age to seed eligible earlier-age non-capitol buildings when creating a city.

Server parity implemented:

- Retail `age.txt` now imports into `RetailRules.Ages`, and each retail advance preserves `AgeID`.
- Runtime rules now project `Ages`, `StartAgeID`, `EndAgeID`, and per-advance `AgeID`.
- `GrantStartingAgeAdvances` grants all advances from ages earlier than the selected start age and initializes a research target from public runtime rules.
- Research legality in engine and public projection suppresses advances after `EndAgeID`.
- Research completion updates `Nation.CurrentAgeID` once the completed advance's age reaches its `NeedAdvances` threshold.
- City founding seeds eligible non-capitol buildings whose enabling advance belongs to an age before the configured/current starting age.

Validation:

- `go test ./internal/gameplay/rules -run TestRetailAdvanceAgesAreProjected -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetailStartingAgeGrantsEarlierAgeAdvances|TestRetailAdvanceCompletionEntersNewAgeAfterNeedAdvances' -count=1`
- `go test ./internal/gameplay/engine -run TestRetailStartingAgeSeedsEarlierAgeCityBuildings -count=1`
- `go test ./internal/gameplay/visibility -run TestProjectionSuppressesResearchBeyondConfiguredEndAge -count=1`
- `go test ./internal/gameplay/rules ./internal/gameplay/engine ./internal/gameplay/visibility ./internal/gameplay/ai -count=1`
