# Task 097 source notes — grouped combat/city-capture edge unit flags verified

Retail source evidence:

- `UnitRecord.cpp:1413-1539,1663-1775` parses `CantCaptureCity`, `SingleUse`, `NuclearAttack`, `VictoryEnslavement`, and `DeathEffectsHappy`; `UnitRecord.h:961-1260` exposes the corresponding accessors.
- `armyevent.cpp:1001-1012` invokes `DoVictoryEnslavement` for victorious armies/defenders, and `armyevent.cpp:1169` checks `IsCantCaptureCity()` before a victorious army can capture a city.
- `ArmyData.cpp:5616-5637` handles all-single-use combat armies and removes single-use units after use.
- `CityEvent.cpp:320-347` and `ArmyData.cpp:5496` apply `DeathEffectsHappy` casualties to happiness/peace-movement style effects.
- `unitutil.cpp:232`, `CityData.cpp:4178`, `ArmyData.cpp:8328-8337`, and `Player.cpp:10202` use `NuclearAttack` as a separate nuke action/eligibility mechanic, not normal capture.
- `Civilian` is parsed/exposed (`UnitRecord.cpp:2176-2177`, `UnitRecord.h:1575`) but retail call sites found here are AI/diplomacy/threat-classification paths (`ArmyData.cpp:7766-7774`, `regardevent.cpp:82-104`, `mapanalysis.cpp:296`), not grouped combat damage/city-capture execution.

Server parity:

- `VictoryEnslavement`, `CantCaptureCity`, `SingleUse`, `DeathEffectsHappy`, and `NuclearAttack` are imported into unit effects and covered by grouped combat, bombard, city capture, peace movement, and nuke action tests.
- `Civilian` is preserved as a unit effect for AI/diplomacy consumers; no separate server combat/city-capture mutation is required by the inspected retail combat/capture call sites.

Validation:

- `go test ./internal/gameplay/engine -run 'TestRetailSingleUseBombarderIsConsumedAfterBombardment|TestRetailSingleUseAttackerIsConsumedAfterCombat|TestRetailVictoryEnslavementAddsPopulationToNearestWinnerCity|TestRetailCantCaptureCityPreventsCityCapture|TestRetailCombatDeathWithDeathEffectsHappyAffectsPeaceMovement|TestRetailCombatDeathWithoutDeathEffectsHappyDoesNotAffectPeaceMovement|TestNuclearAttackUnitNukesForeignCityAndIsConsumed|TestNuclearAttackProtectionConsumesMissileWithoutDamagingCity' -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
